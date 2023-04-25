# benchmarking neo4j

import random
import time
from neo4j import GraphDatabase


URI = "neo4j://localhost:7687"
AUTH = ("neo4j", "12345678")


def create_person(tx, name):
    tx.run("CREATE (a:Person {name: $name})", name=name)


def add_friend(tx, name, friend_name):
    tx.run(
        "MATCH (a:Person {name: $name}) "
        "MATCH (b:Person {name: $friend_name}) "
        "MERGE (a)-[:KNOWS]->(b)",
        name=name,
        friend_name=friend_name,
    )


if __name__ == "__main__":
    g = GraphDatabase.driver(URI, auth=AUTH)
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
        # create neo4j graph

        # add nodes

        for i, vertex in enumerate(vertexes):
            with g.session() as session:
                session.execute_write(create_person, vertex)

        # calculate time taken to add edges

        start = time.time()
        last = start

        for i, edge in enumerate(edges):
            with g.session() as session:
                session.execute_write(add_friend, edge[0], edge[1])
                session.execute_write(add_friend, edge[1], edge[0])
            if i % 1000 == 0:
                print("Added " + str(i) + " edges in time = " + str(time.time() - last))
                last = time.time()

        end = time.time()

        print(
            "Time taken to add "
            + str(len(edges))
            + " edges: "
            + str(end - start)
            + " seconds"
        )

    # delete all nodes and edges

    def delete_all():
        with g.session() as session:
            session.run("MATCH (n) DETACH DELETE n")

    def find_distance_between_two_random_nodes():
        # find distance between two random nodes

        start = time.time()

        for _ in range(0, 100):
            node1, node2 = random.choice(vertexes), random.choice(vertexes)
            if node1 == node2:
                continue
            with g.session() as session:
                r = session.run(
                    "MATCH (a:Person), (b:Person) WHERE a.name = $name1 AND b.name = $name2 MATCH p = shortestPath((a)-[*]-(b)) RETURN length(p)",
                    name1=node1,
                    name2=node2,
                )
                print(
                    "Distance between "
                    + node1
                    + " and "
                    + node2
                    + " is: "
                    + str(r.single()[0])
                )

        end = time.time()

        print("---------------------------------------")
        print(
            "Total time taken to find distance between two random nodes (100 ops): "
            + str(end - start)
            + " seconds"
        )
        print("Time taken per op: " + str((end - start) / 100) + " seconds")

    delete_all()
    # create_graph()
    # find_distance_between_two_random_nodes()

    g.close()
