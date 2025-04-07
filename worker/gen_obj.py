def read_vertices(filename):
    vertices = []
    with open(filename, "r") as f:
        for line in f:
            parts = line.strip().split()
            if not parts:
                continue
            if len(parts) == 2:
                x, y = map(float, parts)
                z = 0.0
            else:
                x, y, z = map(float, parts[:3])
            vertices.append((x, y, z))
    return vertices

def read_edges(filename):
    edges = []
    with open(filename, "r") as f:
        for line in f:
            parts = line.strip().split()
            if len(parts) < 2:
                continue
            u, v = map(int, parts[:2])
            edges.append((u, v))
    return edges

def export_to_obj(vertices, edges, output_path):
    with open(output_path, "w") as f:
        # write vertices
        for v in vertices:
            f.write(f"v {v[0]} {v[1]} {v[2]}\n")

        # write edges as lines
        for u, v in edges:
            # .obj format uses 1-based indexing!
            f.write(f"l {u + 1} {v + 1}\n")

def main():
    import sys
    if len(sys.argv) != 2:
        print("Usage: python export_graph_obj.py <work dir>")
        return

    work_dir = sys.argv[1]
    vertex_file = f"{work_dir}/embedding.txt"
    edge_file   = f"{work_dir}/graph.txt"
    output_file = f"{work_dir}/out.obj"

    vertices = read_vertices(vertex_file)
    edges = read_edges(edge_file)

    export_to_obj(vertices, edges, output_file)
    print(f"Exported .obj to {output_file}")

if __name__ == "__main__":
    main()
