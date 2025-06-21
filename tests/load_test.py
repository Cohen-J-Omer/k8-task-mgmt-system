import time
import requests
import threading

# def load_task():
#     url = "http://localhost:8080/tasks"
#     headers = {"Authorization": "Bearer hardcoded-token"}
#     task = {"title": "LoadTest", "description": "Testing", "completed": False}
#     for _ in range(100):
#         requests.post(url, json=task, headers=headers)

# def load_task():
#     url = "http://localhost:8080/tasks"
#     headers = {"Authorization": "Bearer hardcoded-token"}
#     for _ in range(100):
#         response = requests.get(url, headers=headers)
#         # Optionally print or check the response status
#         print(response.status_code, response.json())


# def load_task():
#     url = "http://localhost:8080/tasks"
#     headers = {"Authorization": "Bearer hardcoded-token"}
#     for _ in range(100):
#         response = requests.get(url, headers=headers)
#         # Optionally print or check the response status
#         print(response.status_code, response.json())

# threads = []
# for _ in range(20):
#     t = threading.Thread(target=load_task)
#     t.start()
#     threads.append(t)

# for t in threads:
#     t.join()



DURATION_SECONDS = 120  # Run for 2 minutes
NUM_THREADS = 50
URL = "http://localhost:8080/tasks"
HEADERS = {"Authorization": "Bearer hardcoded-token"}

def load_task(stop_event):
    while not stop_event.is_set():
        try:
            response = requests.get(URL, headers=HEADERS)
            print(f"Status: {response.status_code}, Body: {response.text}")
        except Exception as e:
            print(f"Request failed: {e}")

stop_event = threading.Event()
threads = []

for _ in range(NUM_THREADS):
    t = threading.Thread(target=load_task, args=(stop_event,))
    t.start()
    threads.append(t)

# Let the threads run for the specified duration
time.sleep(DURATION_SECONDS)
stop_event.set()  # Signal all threads to stop

# Wait for all threads to finish
for t in threads:
    t.join()