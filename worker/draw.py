import numpy as np
import matplotlib.pyplot as plt
import sys

def read_coords(filename):
    coords = []
    with open(filename, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            if "," in line:
                x, y = map(float, line.split(","))
            else:
                x, y = map(float, line.split())
            coords.append((x, y))
    return np.array(coords)

def read_edges(filename):
    edges = []
    with open(filename, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            u, v = map(int, line.split())
            edges.append((u, v))
    return edges

def draw_graph(coords, edges, output_filename="graph_edges_only.png"):
    plt.style.use("seaborn-v0_8-white")

    fig, ax = plt.subplots(figsize=(16, 9), dpi=150)

    for (u, v) in edges:
        if u < len(coords) and v < len(coords):
            x_vals = [coords[u, 0], coords[v, 0]]
            y_vals = [coords[u, 1], coords[v, 1]]
            ax.plot(x_vals, y_vals, color="royalblue", alpha=0.7, linewidth=1.5)

    ax.set_axis_off()
    plt.tight_layout()

    plt.savefig(output_filename, bbox_inches="tight")
    print(f"Graph image with edges only saved as {output_filename}")
    plt.show()

if __name__ == "__main__":
    args = sys.argv
    if len(args) != 1:
        print("Usage: python draw.py <work dir>")
    work_dir = args[1]
    coords = read_coords(f"{work_dir}/embedding.txt")
    edges = read_edges(f"{work_dir}/graph.txt")
    print(f"Read {len(coords)} vertices and {len(edges)} edges.")
    draw_graph(coords, edges, output_filename=f"{work_dir}/graph_edges_only.png")
