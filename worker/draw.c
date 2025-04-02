// TO RUN THIS USE:
/*
 * Compile (Linux):
 *   gcc -std=c99 -O2 -o draw draw.c -lm
 *
 * Run:
 *   ./draw_png
*/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <time.h>

#define GRAPH_PATH "/tmp/graph/graph.txt"
#define EMBED_PATH "/tmp/graph/embedding.txt"
#define OUT_PATH   "/tmp/graph/out.png"

// ---------------------------------------------------------------------
// 1) Include the stb_image_write library in this file.
#define STB_IMAGE_WRITE_IMPLEMENTATION
#include "stb_image_write.h"

// ---------------------------------------------------------------------
// We'll store each vertex's x,y in double precision.
typedef struct {
    double x;
    double y;
} Vertex;

// ---------------------------------------------------------------------
// A basic Bresenham line to draw single-pixel lines into an RGB buffer.
void draw_line(unsigned char *img, int width, int height,
               int x0, int y0, int x1, int y1,
               unsigned char r, unsigned char g, unsigned char b)
{
    // clamp to boundaries
    if(x0 < 0) x0 = 0; if(x0 >= width)  x0 = width - 1;
    if(x1 < 0) x1 = 0; if(x1 >= width)  x1 = width - 1;
    if(y0 < 0) y0 = 0; if(y0 >= height) y0 = height - 1;
    if(y1 < 0) y1 = 0; if(y1 >= height) y1 = height - 1;

    int dx = abs(x1 - x0);
    int sx = (x0 < x1) ? 1 : -1;
    int dy = -abs(y1 - y0);
    int sy = (y0 < y1) ? 1 : -1;
    int err = dx + dy;

    int x = x0, y = y0;
    while(1)
    {
        // invert y to store top->down in memory
        int row = (height - 1 - y);
        int col = x;
        int idx = (row * width + col) * 3; // 3 bytes per pixel (RGB)
        img[idx + 0] = r;
        img[idx + 1] = g;
        img[idx + 2] = b;

        if(x == x1 && y == y1)
            break;
        int e2 = 2*err;
        if(e2 >= dy) { err += dy; x += sx; }
        if(e2 <= dx) { err += dx; y += sy; }
    }
}

// ---------------------------------------------------------------------
// Minimal routines to read vertex coords and edges from text files.
Vertex* read_coords(const char *filename, int *n)
{
    FILE *fp = fopen(filename, "r");
    if(!fp) {
        fprintf(stderr, "Cannot open coords file '%s'\n", filename);
        return NULL;
    }
    int cap = 1000;
    int count = 0;
    Vertex *arr = (Vertex*) malloc(cap * sizeof(Vertex));

    while(!feof(fp))
    {
        double x, y;
        if(fscanf(fp, "%lf %lf", &x, &y) == 2)
        {
            if(count >= cap)
            {
                cap *= 2;
                arr = (Vertex*) realloc(arr, cap * sizeof(Vertex));
            }
            arr[count].x = x;
            arr[count].y = y;
            count++;
        }
        else
        {
            char tmp[256];
            if(!fgets(tmp, 256, fp)) break;
        }
    }
    fclose(fp);
    *n = count;
    return arr;
}

int* read_edges(const char *filename, int *m)
{
    FILE *fp = fopen(filename, "r");
    if(!fp) {
        fprintf(stderr, "Cannot open graph file '%s'\n", filename);
        return NULL;
    }
    int cap = 1000;
    int count = 0;
    int *arr = (int*) malloc(cap * 2 * sizeof(int));

    while(!feof(fp))
    {
        int u, v;
        if(fscanf(fp, "%d %d", &u, &v) == 2)
        {
            if(count >= cap)
            {
                cap *= 2;
                arr = (int*) realloc(arr, cap * 2 * sizeof(int));
            }
            arr[count*2 + 0] = u;
            arr[count*2 + 1] = v;
            count++;
        }
        else
        {
            char tmp[256];
            if(!fgets(tmp, 256, fp)) break;
        }
    }
    fclose(fp);
    *m = count;
    return arr;
}

// ---------------------------------------------------------------------
// Main
int main(int argc, char *argv[])
{
    // Параметры для путей к файлам
    const char *graphFile = GRAPH_PATH;
    const char *coordsFile = EMBED_PATH;
    const char *outFile = OUT_PATH;

    // Обработка аргументов командной строки
    if (argc > 1) graphFile = argv[1];
    if (argc > 2) coordsFile = argv[2];
    if (argc > 3) outFile = argv[3];

    printf("Using files:\n");
    printf("  Graph file: %s\n", graphFile);
    printf("  Embedding file: %s\n", coordsFile);
    printf("  Output file: %s\n", outFile);
    int width = 1200, height = 800;
    // if(scanf("%d %d", &width, &height) != 2) {
    //     fprintf(stderr, "Usage: Provide width and height via standard input (e.g., 1200 800).\n");
    //     return 1;
    // }
    // if(width < 1 || height < 1)
    // {
    //     fprintf(stderr, "Error: invalid image size.\n");
    //     return 1;
    // }

    clock_t start = clock();

    // Read vertex coords
    int n;
    Vertex *verts = read_coords(coordsFile, &n);
    if(!verts || n==0)
    {
        fprintf(stderr, "No vertices read.\n");
        return 1;
    }

    // Read edges
    int m;
    int *edges = read_edges(graphFile, &m);
    if(!edges || m==0)
    {
        fprintf(stderr, "No edges read.\n");
        return 1;
    }

    // Определение степени разреженности графа
    double max_possible_edges = (double)n * (n - 1) / 2;
    double sparsity_factor = max_possible_edges / (m > 0 ? m : 1);
    sparsity_factor = (sparsity_factor > 1000) ? 1000 : sparsity_factor; // ограничение сверху

    // Расчет базового размера для большого разреженного графа
    width = (int)(1000 + 3500 * log10((double)n) / log10(1000000.0));
    height = (int)(800 + 3000 * log10((double)m) / log10(1000000.0));

    // Учитываем разреженность - разреженные графы требуют больше места
    double sparse_multiplier = 1.0 + log10(sparsity_factor) / 2.0;
    width = (int)(width * sparse_multiplier);
    height = (int)(height * sparse_multiplier);

    // Установка разумных пределов
    if (width > 8000) width = 12000;
    if (height > 6000) height = 8000;
    if (width < 1200) width = 1200;
    if (height < 800) height = 800;

    width = 1200;
    height = 800;

    // Find bounding box
    double minx = 1e30, maxx = -1e30;
    double miny = 1e30, maxy = -1e30;
    for(int i=0; i<n; i++)
    {
        if(verts[i].x < minx) minx = verts[i].x;
        if(verts[i].x > maxx) maxx = verts[i].x;
        if(verts[i].y < miny) miny = verts[i].y;
        if(verts[i].y > maxy) maxy = verts[i].y;
    }
    double rangex = (maxx>minx)? (maxx-minx) : 1e-9;
    double rangey = (maxy>miny)? (maxy-miny) : 1e-9;

    // We'll scale x-> [0,width], y-> [0,height], ignoring aspect ratio.
    double scaleX = (double)width  / rangex;
    double scaleY = (double)height / rangey;

    // Allocate an RGB image buffer (3 bytes/pixel) with a white background
    unsigned char *img = (unsigned char*) malloc(width*height*3);
    if(!img)
    {
        fprintf(stderr, "Error allocating image.\n");
        return 1;
    }
    // Fill with white (R=G=B=255)
    memset(img, 255, width*height*3);

    // We'll draw edges in royal blue: (R=65, G=105, B=225).
    unsigned char rr = 65, gg = 105, bb = 225;

    // Draw edges with Bresenham
    for(int e=0; e<m; e++)
    {
        int u = edges[e*2 + 0];
        int v = edges[e*2 + 1];
        if(u<0 || u>=n || v<0 || v>=n) continue;
        double x0 = (verts[u].x - minx)*scaleX;
        double y0 = (verts[u].y - miny)*scaleY;
        double x1 = (verts[v].x - minx)*scaleX;
        double y1 = (verts[v].y - miny)*scaleY;
        int ix0 = (int)round(x0);
        int iy0 = (int)round(y0);
        int ix1 = (int)round(x1);
        int iy1 = (int)round(y1);
        draw_line(img, width, height, ix0, iy0, ix1, iy1, rr, gg, bb);
    }

    // Write as PNG using stb_image_write
    // We store top->down rows, but stbi_write_png expects row0 at top,
    // so no flipping needed. Just pass 'width*3' as stride.
    int stride_in_bytes = width * 3;
    if(!stbi_write_png(outFile, width, height, 3, img, stride_in_bytes))
    {
        fprintf(stderr, "Failed to write PNG '%s'\n", outFile);
    }
    else
    {
        printf("Wrote %s (%dx%d)\n", outFile, width, height);
    }

    free(img);
    free(verts);
    free(edges);

    clock_t end = clock();
    double elapsed = (double)(end - start)/CLOCKS_PER_SEC;
    printf("Elapsed time: %.6f seconds\n", elapsed);

    return 0;
}