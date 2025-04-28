import requests
import numpy as np
import time
import random
import json
from concurrent.futures import ThreadPoolExecutor

# Configuration
API_URL = "http://localhost:8080"
NUM_PERSONS = 5
NUM_TEST_PERSONS = 5
FEATURES_LENGTH = 256
MAX_WORKERS = 20

def generate_random_person():
    """Generate a random person with name and normalized feature vector."""
    # Generate a random name with unique identifier to ensure we can find it later
    unique_id = random.randint(10000, 99999)
    first_names = ["Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Hannah", "Ian", "Julia"]
    last_names = ["Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia", "Rodriguez", "Wilson"]
    name = f"{random.choice(first_names)} {random.choice(last_names)}_{unique_id}"
    
    # Generate random features and normalize them for better cosine similarity
    features = np.random.randn(FEATURES_LENGTH)
    norm = np.linalg.norm(features)
    if norm > 0:
        features = features / norm
    
    return {
        "name": name,
        "features": features.tolist()
    }

def add_person(person):
    """Add a person to the database via API."""
    try:
        response = requests.post(f"{API_URL}/add_person", json=person)
        if response.status_code == 200:
            return True
        else:
            print(f"Failed to add person: {response.status_code} - {response.text}")
            return False
    except Exception as e:
        print(f"Error adding person: {e}")
        return False

def add_persons_in_parallel(persons):
    """Add multiple persons in parallel using thread pool."""
    success_count = 0
    
    print(f"Adding {len(persons)} persons to the database...")
    start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        results = list(executor.map(add_person, persons))
    
    success_count = sum(results)
    failure_count = len(results) - success_count
    
    elapsed_time = time.time() - start_time
    print(f"Added {success_count} persons, {failure_count} failures, took {elapsed_time:.2f} seconds")
    
    return success_count

def find_similar_persons(features, top_n=5):
    """Query the API for similar persons."""
    try:
        # Convert features to query parameters
        feature_params = [f"features={f}" for f in features]
        query_string = "&".join(feature_params) + f"&top_n={top_n}"
        
        response = requests.get(f"{API_URL}/get_similar_person?{query_string}")
        if response.status_code == 200:
            return response.json()
        else:
            print(f"Failed to get similar persons: {response.status_code} - {response.text}")
            return None
    except Exception as e:
        print(f"Error getting similar persons: {e}")
        return None

def test_person_similarity(person):
    """Test if a person is correctly identified as most similar to themselves."""
    print(f"\nTesting search for person: {person['name']}")
    
    # Query the API for similar persons
    results = find_similar_persons(person['features'], top_n=3)
    
    if not results or "persons" not in results:
        print(f"❌ TEST FAILED: No valid results returned from API for {person['name']}")
        return False
    
    api_results = results["persons"]
    if len(api_results) == 0:
        print(f"❌ TEST FAILED: No persons found for {person['name']}")
        return False
    
    # Get the most similar person
    most_similar = api_results[0]["Person"]["Name"]
    similarity_score = api_results[0]["Score"]
    
    print(f"Most similar person: {most_similar}")
    print(f"Similarity score: {similarity_score}")
    
    # Check if the most similar person is the one we queried
    if most_similar == person['name'] and similarity_score > 0.999:
        print(f"✅ TEST PASSED: {person['name']} correctly recognized as most similar to self")
        return True
    else:
        print(f"❌ TEST FAILED: Most similar should be {person['name']}, but got {most_similar}")
        return False

def main():
    print("=" * 80)
    print("Face Recognition API Test")
    print("=" * 80)
    
    # Generate random persons
    print(f"Generating {NUM_PERSONS} random persons...")
    persons = [generate_random_person() for _ in range(NUM_PERSONS)]
    
    # Add persons to the database
    added_count = add_persons_in_parallel(persons)
    
    # if added_count < 100:  # Make sure we have at least 100 persons
    #     print("Too few persons added successfully. Aborting test.")
    #     return
    
    print("\nWaiting a moment for database operations to complete...")
    time.sleep(2)
    
    # Select random persons to test
    test_persons = random.sample(persons, NUM_TEST_PERSONS)
    
    print(f"\nTesting similarity search for {NUM_TEST_PERSONS} random persons...")
    test_results = []
    
    for person in test_persons:
        result = test_person_similarity(person)
        test_results.append(result)
    
    # Summarize results
    success_count = sum(1 for r in test_results if r)
    print("\n" + "=" * 80)
    print(f"Test Summary: {success_count}/{NUM_TEST_PERSONS} tests passed")
    if success_count == NUM_TEST_PERSONS:
        print("✅ ALL TESTS PASSED: The API correctly identifies persons as most similar to themselves")
    else:
        print(f"❌ {NUM_TEST_PERSONS - success_count} TESTS FAILED: The API has issues identifying persons")
    print("=" * 80)

if __name__ == "__main__":
    main()
