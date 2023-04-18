import random
import time

import requests

base_url = "http://0.0.0.0:9000"

n, m = 100, 100

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

# calculate time taken to add edges

start = time.time()
last = start

for i, edge in enumerate(edges):
    requests.post(base_url + "/graph/add-edge", params={"v1": edge[0], "v2": edge[1]})
    if i % 100 == 0:
        print("Added " + str(i) + " edges in time = " + str(time.time() - last))
        last = time.time()
    # time.sleep(1)

end = time.time()

print("Time taken to add edges: " + str(end - start))
