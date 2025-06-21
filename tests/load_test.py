import time
import requests
import threading

DURATION_SECONDS = 120  # Run for 2 minutes
NUM_THREADS = 50  # execute 50 threads
URL = "http://localhost:8080/tasks"
HEADERS = {"Authorization": "Bearer hardcoded-token"}

def load_task(stop_event):
    """
    Continuously sends GET requests to the k8 cluster API service endpoint until the stop_event is set.
    """
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

time.sleep(DURATION_SECONDS)
stop_event.set()  # Signal threads to stop 

for t in threads:
    t.join()