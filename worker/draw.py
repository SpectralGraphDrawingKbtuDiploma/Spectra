import sys
import math
import time
from PIL import Image

def bresenham_line(x0, y0, x1, y1):
    points = []
    dx = abs(x1 - x0)
    dy = abs(y1 - y0)
    sx = 1 if x0 < x1 else -1
    sy = 1 if y0 < y1 else -1
    err = dx - dy

    while True:
        points.append((x0, y0))
        if x0 == x1 and y0 == y1:
            break
        e2 = 2 * err
        if e2 > -dy:
            err -= dy
            x0 += sx
        if e2 < dx:
            err += dx
            y0 += sy
    return points

def read_vertex_coords(coords_file):
    vx = []
    vy = []
    xmin, ymin = float('inf'), float('inf')
    xmax, ymax = float('-inf'), float('-inf')
    with open(coords_file, 'r') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            # If the line contains a comma, split on comma; otherwise, split on whitespace.
            if ',' in line:
                parts = line.split(',')
            else:
                parts = line.split()
            if len(parts) != 2:
                continue
            x = float(parts[0])
            y = float(parts[1])
            vx.append(x)
            vy.append(y)
            xmin = min(xmin, x)
            xmax = max(xmax, x)
            ymin = min(ymin, y)
            ymax = max(ymax, y)
    return vx, vy, xmin, xmax, ymin, ymax

def read_graph_edges(graph_file):

    edges = []
    with open(graph_file, 'r') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            parts = line.split()
            if len(parts) < 2:
                continue
            u = int(parts[0])
            v = int(parts[1])
            edges.append((u, v))
    return edges

def main():
    args = sys.argv
    if len(args) != 2:
        print("Usage: python draw.py <work dir>")
    work_dir = args[1]

    graph_file = work_dir + "/graph.txt"
    coords_file = work_dir + "/embedding.txt"
    output_file = work_dir + "/graph_edges_only.png"

    # Read vertex coordinates
    vx, vy, xmin, xmax, ymin, ymax = read_vertex_coords(coords_file)
    n = len(vx)
    if n == 0:
        print("No vertex coordinates found.")
        sys.exit(1)

    # Compute aspect ratio and image dimensions
    aspect_ratio = (xmax - xmin) / (ymax - ymin) if (ymax - ymin) != 0 else 1
    width = int(math.sqrt(100 * n / aspect_ratio))
    height = int(aspect_ratio * width)
    print("Computed image dimensions: width={}, height={}".format(width, height))

    # Scale vertex coordinates to the image dimensions
    scaled_vx = []
    scaled_vy = []
    for i in range(n):
        sx = int((vx[i] - xmin) / (xmax - xmin) * width)
        sy = int((vy[i] - ymin) / (ymax - ymin) * height)
        sx = min(max(sx, 0), width - 1)
        sy = min(max(sy, 0), height - 1)
        scaled_vx.append(sx)
        scaled_vy.append(sy)

    # Read graph edges
    edges = read_graph_edges(graph_file)
    if not edges:
        print("No edges found in the graph file.")
        sys.exit(1)

    # Create a white RGBA image
    image = Image.new("RGBA", (width, height), "white")
    pixels = image.load()

    # Set the desired edge color (royal blue)
    edge_color = (65, 105, 225, 255)

    # Draw each edge using Bresenham's line algorithm
    for u, v in edges:
        if u < 0 or u >= n or v < 0 or v >= n:
            continue  # Skip invalid indices.
        x0, y0 = scaled_vx[u], scaled_vy[u]
        x1, y1 = scaled_vx[v], scaled_vy[v]
        for x, y in bresenham_line(x0, y0, x1, y1):
            if 0 <= x < width and 0 <= y < height:
                pixels[x, y] = edge_color

    # Save the image as PNG
    image.save(output_file)
    print("Graph image saved to '{}'".format(output_file))

if __name__ == "__main__":
    start = time.perf_counter()
    main()
    end = time.perf_counter()
    print(f"Elapsed time: {end - start:.6f} seconds")
