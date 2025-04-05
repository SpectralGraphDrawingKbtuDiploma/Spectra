#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <map>
#include <set>
#include <algorithm>
#include <Eigen/Dense>
#include <Eigen/Sparse>
#include <Eigen/Eigenvalues>

// Include Spectra for large sparse matrices
#include <Spectra/SymEigsSolver.h>
#include <Spectra/MatOp/SparseSymMatProd.h>

// Define a threshold for using sparse computation
const int SPARSE_THRESHOLD = 1000;

// Function to read graph from file and return Laplacian matrix (dense or sparse)
std::tuple<Eigen::MatrixXd, Eigen::SparseMatrix<double>, std::map<std::string, int>, bool>
readGraph(const std::string& filepath) {
    std::ifstream file(filepath);
    if (!file.is_open()) {
        std::cerr << "Error: Unable to open input file: " << filepath << std::endl;
        exit(1);
    }

    std::set<std::string> nodeSet;
    std::vector<std::pair<std::string, std::string>> edges;

    std::string node1, node2;
    while (file >> node1 >> node2) {
        nodeSet.insert(node1);
        nodeSet.insert(node2);
        edges.push_back({node1, node2});
    }

    // Map node names to indices
    std::map<std::string, int> nodeMap;
    int idx = 0;
    for (const auto& node : nodeSet) {
        nodeMap[node] = idx++;
    }

    int n = nodeSet.size();
    bool useSparse = (n > SPARSE_THRESHOLD);

    if (useSparse) {
        // For sparse matrices
        typedef Eigen::Triplet<double> T;
        std::vector<T> adjTriplets;
        std::vector<T> degTriplets;
        std::vector<int> degrees(n, 0);

        // Construct adjacency matrix
        for (const auto& edge : edges) {
            int i = nodeMap[edge.first];
            int j = nodeMap[edge.second];
            adjTriplets.push_back(T(i, j, 1.0));
            adjTriplets.push_back(T(j, i, 1.0));
            degrees[i]++;
            degrees[j]++;
        }

        // Construct degree matrix
        for (int i = 0; i < n; i++) {
            degTriplets.push_back(T(i, i, degrees[i]));
        }

        // Create sparse matrices
        Eigen::SparseMatrix<double> adjMatrix(n, n);
        Eigen::SparseMatrix<double> degMatrix(n, n);
        adjMatrix.setFromTriplets(adjTriplets.begin(), adjTriplets.end());
        degMatrix.setFromTriplets(degTriplets.begin(), degTriplets.end());

        // Create Laplacian matrix
        Eigen::SparseMatrix<double> laplacian = degMatrix - adjMatrix;

        // Return empty dense matrix as placeholder
        return {Eigen::MatrixXd(), laplacian, nodeMap, useSparse};
    } else {
        // For dense matrices
        Eigen::MatrixXd adjMatrix = Eigen::MatrixXd::Zero(n, n);
        for (const auto& edge : edges) {
            int i = nodeMap[edge.first];
            int j = nodeMap[edge.second];
            adjMatrix(i, j) = 1.0;
            adjMatrix(j, i) = 1.0;  // For undirected graph
        }

        // Create degree matrix
        Eigen::MatrixXd degMatrix = Eigen::MatrixXd::Zero(n, n);
        for (int i = 0; i < n; i++) {
            degMatrix(i, i) = adjMatrix.row(i).sum();
        }

        // Create Laplacian matrix
        Eigen::MatrixXd laplacian = degMatrix - adjMatrix;

        // Return empty sparse matrix as placeholder
        return {laplacian, Eigen::SparseMatrix<double>(), nodeMap, useSparse};
    }
}

int main(int argc, char* argv[]) {
    if (argc != 3) {
        std::cerr << "Usage: " << argv[0] << " input_filepath output_path" << std::endl;
        return 1;
    }

    std::string inputFilepath = argv[1];
    std::string outputPath = argv[2];

    // Ensure output path ends with /
    if (outputPath.back() != '/') {
        outputPath += '/';
    }

    std::string outputFilepath = outputPath + "embedding.txt";

    // Read graph and compute Laplacian
    auto [denseLaplacian, sparseLaplacian, nodeMap, useSparse] = readGraph(inputFilepath);

    int n = useSparse ? sparseLaplacian.rows() : denseLaplacian.rows();

    // Vectors to store the 2nd and 3rd eigenvectors
    Eigen::MatrixXd selectedEigenvectors(n, 2);

    if (useSparse) {
        std::cout << "Using Spectra for large sparse graph with " << n << " nodes" << std::endl;

        // Compute eigenvectors using Spectra for sparse matrices
        Spectra::SparseSymMatProd<double> op(sparseLaplacian);

        // Compute 3 smallest eigenvalues/eigenvectors (we need 2nd and 3rd)
        Spectra::SymEigsSolver<double, Spectra::SMALLEST_ALGE, Spectra::SparseSymMatProd<double>>
            solver(&op, 3, 6);

        // Initialize and compute
        solver.init();
        int nconv = solver.compute();

        if (solver.info() != Spectra::SUCCESSFUL) {
            std::cerr << "Eigenvalue computation failed!" << std::endl;
            return 1;
        }

        // Get eigenvalues and eigenvectors
        Eigen::VectorXd eigenvalues = solver.eigenvalues();
        Eigen::MatrixXd eigenvectors = solver.eigenvectors();

        // Extract 2nd and 3rd eigenvectors
        selectedEigenvectors.col(0) = eigenvectors.col(1);
        selectedEigenvectors.col(1) = eigenvectors.col(2);

    } else {
        std::cout << "Using Eigen for small graph with " << n << " nodes" << std::endl;

        // Compute eigenvalues and eigenvectors using Eigen for dense matrices
        Eigen::SelfAdjointEigenSolver<Eigen::MatrixXd> solver(denseLaplacian);
        if (solver.info() != Eigen::Success) {
            std::cerr << "Eigenvalue computation failed!" << std::endl;
            return 1;
        }

        // Get eigenvectors for the second and third smallest eigenvalues
        Eigen::VectorXd eigenvalues = solver.eigenvalues();
        Eigen::MatrixXd eigenvectors = solver.eigenvectors();

        // Extract 2nd and 3rd eigenvectors
        selectedEigenvectors.col(0) = eigenvectors.col(1);
        selectedEigenvectors.col(1) = eigenvectors.col(2);
    }

    // Sort nodes by their original index if they were numbered
    std::vector<std::pair<std::string, int>> sortedNodes;
    for (const auto& node : nodeMap) {
        sortedNodes.push_back(node);
    }

    // Sort by node name (assuming they are in format "1", "2", etc.)
    std::sort(sortedNodes.begin(), sortedNodes.end(),
              [](const auto& a, const auto& b) {
                  try {
                      return std::stoi(a.first) < std::stoi(b.first);
                  } catch (...) {
                      return a.first < b.first;
                  }
              });

    // Open output file
    std::ofstream outFile(outputFilepath);
    if (!outFile.is_open()) {
        std::cerr << "Error: Unable to open output file: " << outputFilepath << std::endl;
        return 1;
    }

    // Write the second and third eigenvectors
    for (const auto& node : sortedNodes) {
        int idx = node.second;
        outFile << selectedEigenvectors(idx, 0) << " " << selectedEigenvectors(idx, 1) << std::endl;
    }

    std::cout << "Eigenvectors written to " << outputFilepath << std::endl;

    return 0;
}
