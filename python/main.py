import random
from typing import Iterator, Iterable

import numpy as np
import scipy.sparse as sp
import scipy.sparse.linalg as spla
import matplotlib.pyplot as plt


def read_graph(filename):
    """Читает граф из файла и создаёт разреженную матрицу смежности."""
    edges = []
    mp = dict()
    nodes_count = 0
    cnt = dict()

    with open(filename, "r") as f:
        for line in f:
            if line.strip():
                u, v = map(int, line.split())
                edges.append((u, v))
                edges.append((v, u))
                if mp.get(u) is None:
                    mp[u] = nodes_count
                    nodes_count += 1
                if mp.get(v) is None:
                    mp[v] = nodes_count
                    nodes_count += 1
                cnt.setdefault(mp[v], 0)
                cnt.setdefault(mp[u], 0)
                cnt[mp[u]] += 1
                cnt[mp[v]] += 1

    print(nodes_count)
    row_indexes = [mp[u] for u, v in edges] + [i for i in range(nodes_count)]
    col_indexes = [mp[v] for u, v in edges] + [j for j in range(nodes_count)]
    values = [-1 for _ in edges] + [cnt[i] for i in range(nodes_count)]

    print(row_indexes)
    print(col_indexes)
    random.randint(1, 2)

    new_mtx = sp.coo_matrix((values, (row_indexes, col_indexes)), shape=(nodes_count, nodes_count))
    print(new_mtx)
    return new_mtx


def compute_laplacian_eigenvalues(laplacian, k=3):
    """Вычисляет k наименьших собственных значений разреженной матрицы Лапласа."""
    # degrees = np.array(adj_matrix.sum(axis=1)).flatten()
    # laplacian = sp.diags(degrees) - adj_matrix  # L = D - A

    # Находим k наименьших собственных значений
    eigenvalues, vectors = spla.eigsh(laplacian, k=k, which="SM")

    return eigenvalues, vectors


#
# import matplotlib.pyplot as plt
#
#
# def remove_outliers(data, threshold=1.5):
#     Q1 = np.percentile(data, 25)
#     Q3 = np.percentile(data, 75)
#     IQR = Q3 - Q1
#     lower_bound = Q1 - threshold * IQR
#     upper_bound = Q3 + threshold * IQR
#     return (data >= lower_bound) & (data <= upper_bound)


if __name__ == "__main__":
    filename = "p2p-Gnutella06.txt"  # Используем загруженный файл
    read_graph(filename)
    laplacian = read_graph(filename)
    eigenvalues, vectors = compute_laplacian_eigenvalues(laplacian, k=4)
    print(eigenvalues)
    # print(vectors)
    #
    # print(eigenvalues)
    # print(type(eigenvalues))
    print(vectors)
    # v1 = vectors[:, 1]
    # v2 = vectors[:, 2]
    # v3 = vectors[:, 3]
    # print(v1)
    # print(v2)
    # print(v3)
    # # print(len(v1))
    # # print(len(v2))
    # # for i in range(len(v1)):
    # #     v1[i] *= 100
    # # for i in range(len(v2)):
    # #     v2[i] *= 100
    #
    #
    # print(v1)
    # print(v2)
    # mask = remove_outliers(v1) & remove_outliers(v2)
    # # v1_filtered = v1
    # # v2_filtered = v2
    # v1_filtered = v1[mask]
    # v2_filtered = v2[mask]
    #
    #
    # plt.scatter(v1_filtered, v2_filtered, color="r", label="Nodes")
    # cnt = 0
    # for (i, j) in edges:
    #     # if cnt > 100
    #     if mp[i] in mask.nonzero()[0] and mp[j] in mask.nonzero()[0]:
    #         plt.plot([v1[mp[i]], v1[mp[j]]], [v2[mp[i]], v2[mp[j]]], "b-")
    #     cnt += 1
    #
    # plt.xlabel("X")
    # plt.ylabel("Y")
    # plt.legend()
    # plt.grid(True)
    # plt.show()
