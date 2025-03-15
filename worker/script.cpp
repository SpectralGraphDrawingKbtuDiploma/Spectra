#include <Eigen/Sparse>
#include <Eigen/Dense>
#include <iomanip>
#include <Eigen/Eigenvalues>
#include <iostream>
#include <fstream>
#include <sstream>
#include <vector>
#include <queue>
#include <chrono>
#include <cassert>
#include <cstdlib>
#include <cstring>
#include <algorithm>

#ifdef _OPENMP
#include <omp.h>
#endif

using namespace std;
using namespace Eigen;

// Graph structure holds basic CSR arrays as well as fields for coarsening.
typedef struct {
  long n;                     // number of vertices
  long m;                     // number of nonzeros (for undirected graph, m = 2 * #edges)
  unsigned int *rowOffsets;
  unsigned int *adj;
  long n_coarse;              // number of vertices in coarse graph
  long m_coarse;              // number of edges in coarse graph
  unsigned int *rowOffsetsCoarse;
  unsigned int *adjCoarse;
  int *coarseID;              // mapping from fine to coarse vertices
  double *eweights;           // edge weights in coarse graph
} graph_t;

// Utility: convert number to string with fixed precision.
template <typename T>
std::string to_string_with_precision(const T a_value, const int n = 8) {
  std::ostringstream out;
  out << std::setprecision(n) << a_value;
  return out.str();
}

// Comparison function used in sorting coarse edges.
static int vu_cmpfn_inc(const void *a, const void *b) {
  int *av = (int *) a;
  int *bv = (int *) b;
  if (*av > *bv)
    return 1;
  if (*av < *bv)
    return -1;
  if (av[1] > bv[1])
    return 1;
  if (av[1] < bv[1])
    return -1;
  return 0;
}

// ---------------------------------------------------------------------
// SIMPLE COARSENING: Merge vertices via matching until the coarse graph is small.
static int simpleCoarsening(graph_t *g, int coarseningType) {
  if (coarseningType == 0)
    return 0;

  int num_coarsening_rounds_max = 100;
  int coarse_graph_nmax = 1000;

  int *cID = (int *) malloc(g->n * sizeof(int));
  assert(cID != NULL);
  int *toMatch = (int *) malloc(g->n * sizeof(int));
  assert(toMatch != NULL);

  #ifdef _OPENMP
  #pragma omp parallel for
  #endif
  for (long i = 0; i < g->n; i++) {
    cID[i] = i;
    toMatch[i] = 1;
  }

  int coarse_vert_count = g->n;
  int num_rounds = 0;
  while ((coarse_vert_count > coarse_graph_nmax) && (num_rounds < num_coarsening_rounds_max)) {
    num_rounds++;
    int num_matched = 0;
    for (int i = 0; i < g->n; i++) {
      int u = i;
      while (cID[u] != u) {
        cID[u] = cID[cID[u]];
        u = cID[u];
      }
      if (toMatch[u] == 1) {
        for (unsigned int j = g->rowOffsets[u]; j < g->rowOffsets[u+1]; j++) {
          int v = g->adj[j];
          while (v != cID[v]) {
            cID[v] = cID[cID[v]];
            v = cID[v];
          }
          if (v == u)
            continue;
          if (toMatch[v] == 1) {
            if (u < v)
              cID[v] = u;
            else
              cID[u] = v;
            toMatch[u] = toMatch[v] = 0;
            num_matched += 2;
            break;
          }
        }
      }
    }
    int num_unmatched = coarse_vert_count - num_matched;
    int new_coarse_vert_count = num_matched/2 + num_unmatched;
    coarse_vert_count = new_coarse_vert_count;
    for (int i = 0; i < g->n; i++)
      toMatch[i] = 1;
  }

  int *coarse_edges = (int *) malloc(2 * g->m * sizeof(int));
  assert(coarse_edges != NULL);

  // Update coarse IDs.
  for (int i = 0; i < g->n; i++) {
    int u = cID[i];
    while (u != cID[u])
      u = cID[cID[u]];
    cID[i] = u;
  }

  int *vertIDs = (int *) malloc(g->n * sizeof(int));
  assert(vertIDs != NULL);
  for (int i = 0; i < g->n; i++)
    vertIDs[i] = -1;

  int new_id = 0;
  for (int i = 0; i < g->n; i++) {
    if (cID[i] == i)
      vertIDs[i] = new_id++;
  }
  for (int i = 0; i < g->n; i++) {
    if (vertIDs[i] == -1)
      vertIDs[i] = vertIDs[cID[i]];
  }

  long ecount = 0;
  for (int i = 0; i < g->n; i++) {
    int u = vertIDs[i];
    for (unsigned int j = g->rowOffsets[i]; j < g->rowOffsets[i+1]; j++) {
      int v = vertIDs[g->adj[j]];
      coarse_edges[ecount++] = u;
      coarse_edges[ecount++] = v;
    }
  }
  ecount /= 2;
  qsort(coarse_edges, g->m, 2 * sizeof(int), vu_cmpfn_inc);

  int m_coarse = 1;
  int prev_u = coarse_edges[0];
  int prev_v = coarse_edges[1];
  for (int i = 1; i < ecount; i++) {
    int curr_u = coarse_edges[2*i];
    int curr_v = coarse_edges[2*i+1];
    if ((curr_u != prev_u) || (curr_v != prev_v)) {
      m_coarse++;
      prev_u = curr_u;
      prev_v = curr_v;
    }
  }

  double *eweights = (double *) malloc(m_coarse * sizeof(double));
  assert(eweights != NULL);
  for (int i = 0; i < m_coarse; i++)
    eweights[i] = 1.0;

  unsigned int *rowOffsetsCoarse = (unsigned int *) malloc((coarse_vert_count+1) * sizeof(unsigned int));
  assert(rowOffsetsCoarse != NULL);
  for (int i = 0; i <= coarse_vert_count; i++)
    rowOffsetsCoarse[i] = 0;

  unsigned int *adjCoarse = (unsigned int *) malloc(m_coarse * sizeof(unsigned int));
  assert(adjCoarse != NULL);

  m_coarse = 1;
  prev_u = coarse_edges[0];
  prev_v = coarse_edges[1];
  adjCoarse[0] = prev_v;
  rowOffsetsCoarse[prev_u+1]++;
  for (int i = 1; i < ecount; i++) {
    int curr_u = coarse_edges[2*i];
    int curr_v = coarse_edges[2*i+1];
    if ((curr_u != prev_u) || (curr_v != prev_v)) {
      m_coarse++;
      adjCoarse[m_coarse-1] = curr_v;
      rowOffsetsCoarse[curr_u+1]++;
      prev_u = curr_u;
      prev_v = curr_v;
    } else {
      eweights[m_coarse-1] += 1.0;
    }
  }
  for (int i = 1; i <= coarse_vert_count; i++)
    rowOffsetsCoarse[i] += rowOffsetsCoarse[i-1];

  free(coarse_edges);
  free(cID);
  free(toMatch);

  g->coarseID = vertIDs;
  g->n_coarse = coarse_vert_count;
  g->m_coarse = m_coarse;
  g->eweights = eweights;
  g->adjCoarse = adjCoarse;
  g->rowOffsetsCoarse = rowOffsetsCoarse;

  return 0;
}

// ---------------------------------------------------------------------
// LOAD THE GRAPH INTO AN EIGEN SPARSE MATRIX.
// When coarsening is off (type 0) we build the matrix from the fine graph.
// Otherwise we use the coarse graph arrays.
static int loadToMatrix(SparseMatrix<double,RowMajor>& M, VectorXd& degrees, graph_t *g, int coarseningType) {
  typedef Triplet<double> T;
  vector<T> tripletList;

  if (coarseningType == 0) {
    tripletList.reserve(g->m);
    for (int i = 0; i < g->n; i++) {
      tripletList.push_back(T(i, i, 0.5));
      degrees(i) = g->rowOffsets[i+1] - g->rowOffsets[i];
      double nzv = 1.0 / (2.0 * (g->rowOffsets[i+1] - g->rowOffsets[i]));
      for (unsigned int j = g->rowOffsets[i]; j < g->rowOffsets[i+1]; j++) {
        unsigned int v = g->adj[j];
        tripletList.push_back(T(i, v, nzv));
      }
    }
    M.setFromTriplets(tripletList.begin(), tripletList.end());
  } else {
    for (int i = 0; i < g->n_coarse; i++) {
      double degree_i = 0;
      for (unsigned int j = g->rowOffsetsCoarse[i]; j < g->rowOffsetsCoarse[i+1]; j++)
        degree_i += g->eweights[j];
      degrees(i) = degree_i;
    }
    tripletList.reserve(g->m_coarse);
    for (long i = 0; i < g->n_coarse; i++) {
      double diag_val = 0;
      double inv_2deg = 1.0/(2.0 * degrees(i));
      for (unsigned int j = g->rowOffsetsCoarse[i]; j < g->rowOffsetsCoarse[i+1]; j++) {
        unsigned int v = g->adjCoarse[j];
        if (v == (unsigned int) i)
          diag_val = g->eweights[j] * inv_2deg;
        else
          tripletList.push_back(T(i, v, g->eweights[j] * inv_2deg));
      }
      tripletList.push_back(T(i, i, diag_val + 0.5));
    }
    M.setFromTriplets(tripletList.begin(), tripletList.end());
  }
  return 0;
}

// ---------------------------------------------------------------------
// A simple breadth-first search from a starting vertex. Returns a vector of distances.
static VectorXd bfs(unsigned int *row, unsigned int *col, long N, long M, unsigned int start) {
  VectorXd distances = VectorXd::Constant(N, -1);
  vector<int> visited(N, 0);
  queue<unsigned int> Q;
  Q.push(start);
  visited[start] = 1;
  distances(start) = 0;
  while (!Q.empty()) {
    unsigned int cur = Q.front();
    Q.pop();
    for (unsigned int j = row[cur]; j < row[cur+1]; j++) {
      unsigned int nb = col[j];
      if (!visited[nb]) {
        visited[nb] = 1;
        distances(nb) = distances(cur) + 1;
        Q.push(nb);
      }
    }
  }
  return distances;
}

// ---------------------------------------------------------------------
// HIGH-DIMENSIONAL EMBEDDING (HDE) Initialization.
// It repeatedly computes distance vectors (via BFS) and then performs D-orthogonalization.
// The final two vectors (after an eigen–decomposition) are used as initial second and third eigenvectors.
static int HDE(SparseMatrix<double,RowMajor>& M, graph_t *g, VectorXd& degrees, VectorXd& secondVec, VectorXd& thirdVec) {
  auto startTimer = chrono::high_resolution_clock::now();
  long n = g->n;
  typedef Triplet<double> T;
  vector<T> LTripletList;
  LTripletList.reserve(g->m);
  for (int i = 0; i < g->n; i++) {
    LTripletList.push_back(T(i, i, degrees(i)));
    for (unsigned int j = g->rowOffsets[i]; j < g->rowOffsets[i+1]; j++) {
      unsigned int v = g->adj[j];
      LTripletList.push_back(T(i, v, -1.0));
    }
  }
  SparseMatrix<double,RowMajor> L(n, n);
  L.setFromTriplets(LTripletList.begin(), LTripletList.end());
  auto endTimer = chrono::high_resolution_clock::now();
  cout << "Laplacian load time: " << chrono::duration<double>(endTimer - startTimer).count() << " s." << endl;

  startTimer = chrono::high_resolution_clock::now();
  VectorXi min_dist = VectorXi::Constant(g->n, INT_MAX);
  int maxM = 50;
  MatrixXd dist = MatrixXd::Zero(n, maxM+1);
  MatrixXd dist_bak = MatrixXd::Zero(n, maxM);
  VectorXd init_vec = VectorXd::Ones(n);
  init_vec.normalize();
  dist.col(0) = init_vec;
  int start_idx = 0;
  for (int run_count = 1; run_count <= maxM; run_count++) {
    dist.col(run_count) = bfs(g->rowOffsets, g->adj, n, g->m, start_idx);
    for (int i = 0; i < n; i++) {
      if (dist(i, run_count) < min_dist(i))
        min_dist(i) = (int) dist(i, run_count);
    }
    int max_val = -1;
    for (int i = 0; i < n; i++) {
      if (min_dist(i) > max_val) {
        max_val = min_dist(i);
        start_idx = i;
      }
    }
    dist.col(run_count).normalize();
  }
  int j = 1;
  for (int run_count = 0; run_count < maxM; run_count++) {
    for (int k = 0; k < j; k++) {
      VectorXd dnormvec = dist.col(k).cwiseProduct(degrees);
      double multplr = dist.col(j).dot(dnormvec);
      dist.col(j) = dist.col(j) - (multplr / dnormvec.dot(dist.col(k))) * dist.col(k);
    }
    double normdist = dist.col(j).norm();
    if (normdist < 0.001) {
      cout << "Discarding vector " << j << ", norm: " << normdist << endl;
      j--;
    } else {
      dist.col(j).normalize();
    }
    dist_bak.col(j-1) = dist.col(j);
    j++;
  }
  MatrixXd LX = L * dist_bak;
  MatrixXd XtLX = dist_bak.transpose() * LX;
  SelfAdjointEigenSolver<MatrixXd> es(XtLX);
  MatrixXd init_vecs = dist_bak * es.eigenvectors().leftCols(2).real();
  auto endTimer2 = chrono::high_resolution_clock::now();
  cout << "HDE Initialization time: " << chrono::duration<double>(endTimer2 - startTimer).count() << " s." << endl;
  secondVec = init_vecs.col(0);
  thirdVec  = init_vecs.col(1);
  return 0;
}

// ---------------------------------------------------------------------
// Koren's Power–Iteration Algorithm for computing the second and third eigenvectors.
// It performs D–orthonormalization against previously computed eigenvectors.
static int powerIterationKoren(SparseMatrix<double,RowMajor>& M, VectorXd& degrees, double eps,
                                VectorXd& firstVec, VectorXd& secondVec, VectorXd& thirdVec, int coarseningType) {
  cout << "Using eps " << eps << " for second eigenvector" << endl;
  int n = M.rows();
  VectorXd uk_hat = secondVec;
  VectorXd uk(n);
  VectorXd firstVecD = firstVec.cwiseProduct(degrees);
  double mult1_denom = firstVec.dot(firstVecD);
  VectorXd residual(n);
  int num_iterations1 = 0;
  auto startTimer = chrono::high_resolution_clock::now();
  while (true) {
    uk = uk_hat;
    double mult1_num = uk.dot(firstVecD);
    uk = uk - (mult1_num / mult1_denom) * firstVec;
    uk_hat = M * uk;
    uk_hat.normalize();
    num_iterations1++;
    residual = uk - uk_hat;
    if (residual.norm() < eps)
      break;
  }
  cout << "Num iterations for second eigenvector: " << num_iterations1 << endl;
  secondVec = uk_hat;
  auto endTimer = chrono::high_resolution_clock::now();
  cout << "Second eigenvector computation time: " << chrono::duration<double>(endTimer - startTimer).count() << " s." << endl;

  eps = 2.0 * eps;
  cout << "Using eps " << eps << " for third eigenvector" << endl;
  VectorXd secondVecD = secondVec.cwiseProduct(degrees);
  double mult2_denom = secondVec.dot(secondVecD);
  uk_hat = thirdVec;
  int num_iterations2 = 0;
  startTimer = chrono::high_resolution_clock::now();
  while (true) {
    uk = uk_hat;
    double mult1_num = uk.dot(firstVecD);
    uk = uk - (mult1_num / mult1_denom) * firstVec;
    double mult2_num = uk.dot(secondVecD);
    uk = uk - (mult2_num / mult2_denom) * secondVec;
    uk_hat = M * uk;
    uk_hat.normalize();
    num_iterations2++;
    residual = uk - uk_hat;
    if (residual.norm() < eps)
      break;
  }
  cout << "Num iterations for third eigenvector: " << num_iterations2 << endl;
  thirdVec = uk_hat;
  auto endTimer2 = chrono::high_resolution_clock::now();
  cout << "Third eigenvector computation time: " << chrono::duration<double>(endTimer2 - startTimer).count() << " s." << endl;
  cout << "Dot products of eigenvectors: " << firstVec.dot(secondVec) << " " << firstVec.dot(thirdVec) << " " << secondVec.dot(thirdVec) << endl;
  return 0;
}

// ---------------------------------------------------------------------
// Tutte Refinement: Multiply the coordinate vectors repeatedly with a modified matrix.
static int RefineTutte(SparseMatrix<double,RowMajor>& M, VectorXd& secondVec, VectorXd& thirdVec, int numSmoothing) {
  cout << "Number of smoothing rounds: " << numSmoothing << endl;
  auto startTimer = chrono::high_resolution_clock::now();
  SparseMatrix<double,RowMajor> M2 = 2 * M;
  M2.diagonal().setZero();
  for (int i = 0; i < numSmoothing; i++) {
    secondVec = M2 * secondVec;
    thirdVec = M2 * thirdVec;
  }
  auto endTimer = chrono::high_resolution_clock::now();
  cout << "RefineTutte Time: " << chrono::duration<double>(endTimer - startTimer).count() << " s." << endl;
  return 0;
}

// ---------------------------------------------------------------------
// Write the computed 2D coordinates (using second and third eigenvectors) to an output file.
static int writeCoords(SparseMatrix<double,RowMajor>& M, VectorXd& firstVec, VectorXd& secondVec, VectorXd& thirdVec,
                       int coarseningType, int doHDE, int refineType, double eps, const char *inputFilename, std::string input_path) {
  string outFilename = input_path + "/embedding.txt";
  ofstream fout(outFilename);
  if (!fout.is_open()) {
    cerr << "Error: Cannot open output file " << outFilename << endl;
    return 1;
  }
  int n = M.cols();
  for (int i = 0; i < n; i++) {
    fout << secondVec(i) << " " << thirdVec(i) << "\n";
  }
  fout.close();
  cout << "Embedding written to " << outFilename << endl;
  return 0;
}

// ---------------------------------------------------------------------
// Read graph from a text file (each line: "u v") and build a CSR structure.
graph_t* readGraphFromTxt(const char *filename) {
  ifstream fin(filename);
  if (!fin.is_open()) {
    cerr << "Error: cannot open file " << filename << endl;
    exit(1);
  }
  vector< pair<unsigned int, unsigned int> > edges;
  unsigned int u, v;
  unsigned int maxVertex = 0;
  while (fin >> u >> v) {
    edges.push_back({u, v});
    maxVertex = max(maxVertex, max(u, v));
  }
  fin.close();
  long n = maxVertex + 1;       // assuming vertices are zero-indexed
  long m = edges.size() * 2;    // undirected graph: add both (u,v) and (v,u)

  vector< vector<unsigned int> > adjList(n);
  for (auto &e : edges) {
    u = e.first;
    v = e.second;
    adjList[u].push_back(v);
    adjList[v].push_back(u);
  }

  unsigned int *rowOffsets = (unsigned int *) malloc((n+1) * sizeof(unsigned int));
  assert(rowOffsets != NULL);
  unsigned int *adj = (unsigned int *) malloc(m * sizeof(unsigned int));
  assert(adj != NULL);

  rowOffsets[0] = 0;
  for (long i = 0; i < n; i++) {
    rowOffsets[i+1] = rowOffsets[i] + adjList[i].size();
  }
  for (long i = 0; i < n; i++) {
    for (size_t j = 0; j < adjList[i].size(); j++) {
      adj[rowOffsets[i] + j] = adjList[i][j];
    }
  }

  graph_t *g = (graph_t *) malloc(sizeof(graph_t));
  assert(g != NULL);
  g->n = n;
  g->m = m;
  g->rowOffsets = rowOffsets;
  g->adj = adj;
  g->n_coarse = 0;
  g->m_coarse = 0;
  g->rowOffsetsCoarse = NULL;
  g->adjCoarse = NULL;
  g->coarseID = NULL;
  g->eweights = NULL;

  return g;
}

// ---------------------------------------------------------------------
// MAIN: Parse command-line options, read the graph, run coarsening (if any),
// then compute the spectral embedding using HDE, Koren's algorithm and/or Tutte refinement.
// Finally, write the output embedding.
int main(int argc, char **argv) {
  if (argc != 6) {
    cout << "Usage: " << argv[0] << " <graph.txt> <0/1/2> <0/1> <0/1/2/3>" << endl;
    cout << "    where <graph.txt> is a text file with lines: \"u v\"" << endl;
    cout << "    <0/1/2>: coarsening type (0: none, 1: coarsen and continue, 2: coarsen and stop)" << endl;
    cout << "    <0/1>: HDE flag (0: off, 1: on)" << endl;
    cout << "    <0/1/2/3>: refinement (0: none, 1: Koren, 2: Tutte, 3: Koren+Tutte)" << endl;
    return 1;
  }

  const char *inputFilename = argv[1];
  std::string input_path(argv[5]);
  int coarseningType = atoi(argv[2]);
  int doHDE = atoi(argv[3]);
  int refineType = atoi(argv[4]);

  if (coarseningType == 1)
    cout << "Coarsening graph and continuing" << endl;
  else if (coarseningType == 2)
    cout << "Coarsening graph and stopping" << endl;
  else
    coarseningType = 0;

  if (doHDE)
    cout << "Running High-dimensional embedding (HDE)" << endl;
  else
    doHDE = 0;

  if (refineType == 0)
    cout << "No eigenvector refinement" << endl;
  else if (refineType == 1)
    cout << "Using Koren's algorithm" << endl;
  else if (refineType == 2)
    cout << "Using Tutte refinement" << endl;
  else if (refineType == 3)
    cout << "Using Koren's algorithm followed by Tutte refinement" << endl;

  auto startTimer = chrono::high_resolution_clock::now();

  cout << "Reading graph from file: " << inputFilename << endl;
  graph_t *g = readGraphFromTxt(inputFilename);
  cout << "Graph: vertices = " << g->n << ", edges = " << g->m/2 << endl;

  // Perform coarsening if selected.
  simpleCoarsening(g, coarseningType);

  VectorXd secondCoarse, thirdCoarse;
  if (coarseningType > 0) {
    int n_coarse = g->n_coarse;
    SparseMatrix<double,RowMajor> Mc(n_coarse, n_coarse);
    VectorXd degreesc(n_coarse);
    degreesc.setZero();
    loadToMatrix(Mc, degreesc, g, coarseningType);
    VectorXd firstCoarse = VectorXd::Ones(n_coarse);
    firstCoarse.normalize();
    secondCoarse = VectorXd::Random(n_coarse);
    if (secondCoarse(0) < 0)
      secondCoarse = -secondCoarse;
    secondCoarse.normalize();
    thirdCoarse = VectorXd::Random(n_coarse);
    if (thirdCoarse(0) < 0)
      thirdCoarse = -thirdCoarse;
    thirdCoarse.normalize();
    double epsc = 1e-9;
    powerIterationKoren(Mc, degreesc, epsc, firstCoarse, secondCoarse, thirdCoarse, coarseningType);
    if (coarseningType == 2) {
      writeCoords(Mc, firstCoarse, secondCoarse, thirdCoarse, coarseningType, doHDE, refineType, epsc, inputFilename, input_path);
      return 0;
    }
  }

  // Load the full (fine) graph.
  SparseMatrix<double,RowMajor> M(g->n, g->n);
  VectorXd degrees(g->n);
  degrees.setZero();
  loadToMatrix(M, degrees, g, 0);

  VectorXd firstVec = VectorXd::Ones(g->n);
  firstVec.normalize();
  VectorXd secondVec(g->n);
  VectorXd thirdVec(g->n);

  if (coarseningType == 1) {
    for (long i = 0; i < g->n; i++) {
      secondVec(i) = secondCoarse[g->coarseID[i]];
      thirdVec(i)  = thirdCoarse[g->coarseID[i]];
    }
    secondVec.normalize();
    thirdVec.normalize();
  } else if (doHDE == 1) {
    HDE(M, g, degrees, secondVec, thirdVec);
  } else {
    secondVec = VectorXd::Random(g->n);
    if (secondVec(0) < 0)
      secondVec = -secondVec;
    secondVec.normalize();
    thirdVec = VectorXd::Random(g->n);
    if (thirdVec(0) < 0)
      thirdVec = -thirdVec;
    thirdVec.normalize();
  }

  if (coarseningType != 2) {
    int numTutteSmoothing = 500;
    double eps = 1e-5;
    if (refineType == 0) {
      secondVec.normalize();
      thirdVec.normalize();
    } else if (refineType == 1) {
      powerIterationKoren(M, degrees, eps, firstVec, secondVec, thirdVec, 0);
    } else if (refineType == 2) {
      RefineTutte(M, secondVec, thirdVec, numTutteSmoothing);
    } else if (refineType == 3) {
      RefineTutte(M, secondVec, thirdVec, numTutteSmoothing);
      powerIterationKoren(M, degrees, eps, firstVec, secondVec, thirdVec, 0);
    }
    writeCoords(M, firstVec, secondVec, thirdVec, coarseningType, doHDE, refineType, 0, inputFilename, input_path);
  }

  free(g->rowOffsets);
  free(g->adj);
  if (coarseningType > 0) {
    free(g->rowOffsetsCoarse);
    free(g->adjCoarse);
    free(g->coarseID);
    free(g->eweights);
  }
  free(g);

  auto endTimer = chrono::high_resolution_clock::now();
  cout << "Overall time: " << chrono::duration<double>(endTimer - startTimer).count() << " s." << endl;
  return 0;
}
