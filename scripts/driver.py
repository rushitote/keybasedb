import random
import time

import requests


base_url = "http://0.0.0.0:9000"

n, m = 500, 10

# Generate list of vertexes
vertexes = []
for i in range(0, n):
    vertexes.append("ver" + str(i))

# Generate list of random edges
edges = []
for i in range(0, n - 1):
    for j in range(0, m):
        # randomly choose a vertex from (i + 1) to 1000
        rand = random.randint(i + 1, n - 1)
        edges.append((vertexes[i], vertexes[rand]))

edges = list(set(edges))

print("Number of edges: " + str(len(edges)))


def create_graph():
    # recon graph

    requests.post(base_url + "/graph/recon")

    # calculate time taken to add edges

    start = time.time()
    last = start

    for i, edge in enumerate(edges):
        requests.post(
            base_url + "/graph/add-edge", params={"v1": edge[0], "v2": edge[1]}
        )
        if i % 1000 == 0:
            print("Added " + str(i) + " edges in time = " + str(time.time() - last))
            last = time.time()
        # time.sleep(1)

    end = time.time()

    print(
        "Time taken to add "
        + str(len(edges))
        + " edges: "
        + str(end - start)
        + " seconds"
    )


def find_distance_between_two_random_nodes():
    # find distance between two random nodes

    start = time.time()

    for _ in range(0, 100):
        v1, v2 = random.choice(vertexes), random.choice(vertexes)

        r = requests.post(
            base_url + "/graph/get-degrees",
            params={"v1": v1, "v2": v2},
        )

        end = time.time()

        print("Distance between " + v1 + " and " + v2 + " is " + str(r.text))

    print("---------------------------------------")
    print(
        "Total time taken to find distance between two random nodes (100 ops): "
        + str(end - start)
        + " seconds"
    )
    print("Time taken per op: " + str((end - start) / 100) + " seconds")


# create_graph()
find_distance_between_two_random_nodes()
# requests.post(base_url + "/graph/recon")
