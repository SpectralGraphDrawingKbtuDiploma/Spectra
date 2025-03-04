import numpy as np
import networkx as nx
import sys


def read_graph(filename):
    """
    Читает граф из файла. Файл должен содержать рёбра в формате:
    u v
    где u и v - индексы вершин (нумерация с 0 или 1).
    """
    G = nx.read_edgelist(filename, nodetype=int)  # Читаем список рёбер
    G = nx.convert_node_labels_to_integers(G)  # Перенумеровываем узлы (на случай пропусков)
    G.to_undirected_class()
    return G


def compute_laplacian_eigenvalues(G, k=3):
    """
    Вычисляет k наименьших собственных значений матрицы Лапласа.
    """
    # Получаем плотную матрицу Лапласа
    L = nx.laplacian_matrix(G).toarray()

    # Находим k наименьших собственных значений
    eigenvalues, _ = np.linalg.eigh(L)

    return eigenvalues[:k]


if __name__ == "__main__":
    # if len(sys.argv) < 2:
    #     print("Использование: python script.py graph.txt")
    #     sys.exit(1)

    filename = "p2p-Gnutella06.txt"

    G = read_graph(filename)  # Читаем граф
    eigenvalues = compute_laplacian_eigenvalues(G, k=3)  # Считаем собственные значения

    # Вычисляем метрики графа
    num_nodes = G.number_of_nodes()
    num_edges = G.number_of_edges()
    num_components = nx.number_connected_components(G)
    avg_degree = np.mean([deg for _, deg in G.degree()])

    print("Графовые метрики:")
    print(f"Количество вершин: {num_nodes}")
    print(f"Количество рёбер: {num_edges}")
    print(f"Количество компонент связности: {num_components}")
    print(f"Средняя степень вершины: {avg_degree:.2f}")

    print("Три наименьших собственных значения:", eigenvalues)
