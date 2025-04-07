#include <vtkActor.h>
#include <vtkAxesActor.h>
#include <vtkCellArray.h>
#include <vtkInteractorStyleTrackballCamera.h>
#include <vtkLine.h>
#include <vtkMatrix4x4.h>
#include <vtkOBJExporter.h>
#include <vtkOrientationMarkerWidget.h>
#include <vtkPoints.h>
#include <vtkPolyData.h>
#include <vtkPolyDataMapper.h>
#include <vtkProperty.h>
#include <vtkRenderer.h>
#include <vtkRenderWindow.h>
#include <vtkRenderWindowInteractor.h>
#include <vtkSmartPointer.h>
#include <vtkTransform.h>

#include <fstream>
#include <iostream>
#include <string>
#include <vector>

// Структура для хранения начальных параметров камеры
struct CameraInitialParams {
    double position[3];
    double focalPoint[3];
    double viewUp[3];
};

// Переопределение класса vtkInteractorStyleTrackballCamera для обработки событий клавиатуры
class CustomInteractorStyle : public vtkInteractorStyleTrackballCamera
{
public:
    static CustomInteractorStyle* New();
    vtkTypeMacro(CustomInteractorStyle, vtkInteractorStyleTrackballCamera);

    CustomInteractorStyle() = default;

    void SetCompositeTransform(vtkTransform* transform) {
        this->CompositeTransform = transform;
    }

    void SetStartPos(double x, double y, double z) {
        this->StartPos[0] = x;
        this->StartPos[1] = y;
        this->StartPos[2] = z;
    }

    void SetHUDAxes(vtkAxesActor* axes) {
        this->HUDAxes = axes;
    }

    void SetCamera(vtkCamera* camera) {
        this->Camera = camera;
    }

    void SetCameraInitial(const CameraInitialParams& initialParams) {
        this->CameraInitial = initialParams;
    }

    void OnKeyPress() override {
        // Get the keypress
        vtkRenderWindowInteractor* rwi = this->Interactor;
        std::string key = rwi->GetKeySym();

        double angleStep = 5.0; // degrees per key press
        bool rotated = false;

        if (key == "Up") {
            this->CompositeTransform->RotateX(angleStep);
            rotated = true;
        }
        else if (key == "Down") {
            this->CompositeTransform->RotateX(-angleStep);
            rotated = true;
        }
        else if (key == "Left") {
            this->CompositeTransform->RotateY(angleStep);
            rotated = true;
        }
        else if (key == "Right") {
            this->CompositeTransform->RotateY(-angleStep);
            rotated = true;
        }
        else if (key == "space") {
            // Reset composite transform
            this->CompositeTransform->Identity();
            this->CompositeTransform->Translate(StartPos[0], StartPos[1], StartPos[2]);

            // Restore the camera's initial parameters
            this->Camera->SetPosition(CameraInitial.position);
            this->Camera->SetFocalPoint(CameraInitial.focalPoint);
            this->Camera->SetViewUp(CameraInitial.viewUp);

            // Reset the camera clipping range
            vtkRenderer* renderer = this->Interactor->GetRenderWindow()->GetRenderers()->GetFirstRenderer();
            renderer->ResetCameraClippingRange();

            rotated = true;
        }
        else {
            // Call parent's OnKeyPress if not handled here
            vtkInteractorStyleTrackballCamera::OnKeyPress();
        }

        if (rotated) {
            UpdateHUDAxes();
            this->Interactor->GetRenderWindow()->Render();
        }
    }

    void UpdateHUDAxes() {
        vtkNew<vtkMatrix4x4> mat;
        this->CompositeTransform->GetMatrix(mat);

        // Zero out translation components (we want only rotation in the HUD)
        mat->SetElement(0, 3, 0.0);
        mat->SetElement(1, 3, 0.0);
        mat->SetElement(2, 3, 0.0);

        this->HUDAxes->SetUserMatrix(mat);
    }

private:
    vtkTransform* CompositeTransform = nullptr;
    double StartPos[3] = {0.0, 0.0, 0.0};
    vtkAxesActor* HUDAxes = nullptr;
    vtkCamera* Camera = nullptr;
    CameraInitialParams CameraInitial;
};

vtkStandardNewMacro(CustomInteractorStyle);

// Функция для чтения вершин из файла
std::vector<std::array<double, 3>> read_vertices(const std::string& filename) {
    std::vector<std::array<double, 3>> vertices;
    std::ifstream file(filename);

    if (!file.is_open()) {
        std::cerr << "Error opening file: " << filename << std::endl;
        return vertices;
    }

    std::string line;
    while (std::getline(file, line)) {
        std::istringstream iss(line);
        double x, y, z = 0.0;

        if (!(iss >> x >> y)) {
            continue; // Skip if we couldn't read at least x and y
        }

        // Try to read z, but it's optional
        iss >> z;

        vertices.push_back({x, y, z});
    }

    return vertices;
}

// Функция для чтения рёбер из файла
std::vector<std::array<int, 2>> read_edges(const std::string& filename) {
    std::vector<std::array<int, 2>> edges;
    std::ifstream file(filename);

    if (!file.is_open()) {
        std::cerr << "Error opening file: " << filename << std::endl;
        return edges;
    }

    std::string line;
    while (std::getline(file, line)) {
        std::istringstream iss(line);
        int u, v;

        if (!(iss >> u >> v)) {
            continue; // Skip if we couldn't read both vertices
        }

        edges.push_back({u, v});
    }

    return edges;
}

int main(int argc, char* argv[]) {
    if (argc != 2) {
        std::cerr << "Usage: " << argv[0] << " <work dir>" << std::endl;
        return 1;
    }

    std::string work_dir = argv[1];

    // File paths
    std::string vertex_file = work_dir + "/embedding.txt";
    std::string edge_file = work_dir + "/graph.txt";

    // Read data
    auto vertices = read_vertices(vertex_file);
    auto edges = read_edges(edge_file);

    std::cout << "Loaded " << vertices.size() << " vertices" << std::endl;
    std::cout << "Loaded " << edges.size() << " edges" << std::endl;

    if (vertices.empty()) {
        std::cerr << "No vertices loaded!" << std::endl;
        return 1;
    }

    // Create vtkPoints and add vertices
    vtkNew<vtkPoints> vtk_points;
    for (const auto& pt : vertices) {
        vtk_points->InsertNextPoint(pt[0], pt[1], pt[2]);
    }

    // Create a vtkCellArray for the edges
    vtkNew<vtkCellArray> vtk_lines;
    for (const auto& edge : edges) {
        int u = edge[0];
        int v = edge[1];

        if (u < 0 || v < 0 || u >= vertices.size() || v >= vertices.size()) {
            continue;
        }

        vtkNew<vtkLine> line;
        line->GetPointIds()->SetId(0, u);
        line->GetPointIds()->SetId(1, v);
        vtk_lines->InsertNextCell(line);
    }

    // Create polydata to hold the graph
    vtkNew<vtkPolyData> polyData;
    polyData->SetPoints(vtk_points);
    polyData->SetLines(vtk_lines);

    // Create a mapper and actor for the graph
    vtkNew<vtkPolyDataMapper> mapper;
    mapper->SetInputData(polyData);

    vtkNew<vtkActor> actor;
    actor->SetMapper(mapper);
    actor->GetProperty()->SetColor(65.0/255.0, 105.0/255.0, 225.0/255.0); // royal blue
    actor->GetProperty()->SetLineWidth(2);

    // Create a composite transform and apply it only to the graph actor
    vtkNew<vtkTransform> compositeTransform;
    double startPos[3] = {0.0, 0.0, 0.0};
    compositeTransform->Translate(startPos);
    actor->SetUserTransform(compositeTransform);

    // Create renderer, render window, and interactor
    vtkNew<vtkRenderer> renderer;
    // Set background to a white-grey color
    renderer->SetBackground(0.95, 0.95, 0.95);

    vtkNew<vtkRenderWindow> renderWindow;
    renderWindow->AddRenderer(renderer);

    vtkNew<vtkRenderWindowInteractor> interactor;
    interactor->SetRenderWindow(renderWindow);

    // Add the graph actor
    renderer->AddActor(actor);

    // Create a separate axes actor for the orientation marker (HUD)
    vtkNew<vtkAxesActor> hudAxes;

    // Reset the camera to include all actors
    renderer->ResetCamera();
    renderWindow->Render();

    // Get the active camera and store its initial parameters
    vtkCamera* camera = renderer->GetActiveCamera();

    CameraInitialParams camera_initial;
    camera->GetPosition(camera_initial.position);
    camera->GetFocalPoint(camera_initial.focalPoint);
    camera->GetViewUp(camera_initial.viewUp);

    // Instantiate the custom interactor style with necessary parameters
    vtkNew<CustomInteractorStyle> style;
    style->SetCompositeTransform(compositeTransform);
    style->SetStartPos(startPos[0], startPos[1], startPos[2]);
    style->SetHUDAxes(hudAxes);
    style->SetCamera(camera);
    style->SetCameraInitial(camera_initial);

    interactor->SetInteractorStyle(style);

    // Add an orientation marker widget to display the HUD axes
    vtkNew<vtkOrientationMarkerWidget> orientationWidget;
    orientationWidget->SetOrientationMarker(hudAxes);
    orientationWidget->SetInteractor(interactor);
    orientationWidget->SetViewport(0.0, 0.0, 0.2, 0.2);
    orientationWidget->EnabledOn();
    orientationWidget->InteractiveOff();

    renderWindow->SetSize(800, 600);
    renderWindow->SetWindowName("3D Graph with Rotating Coordinates (VTK - C++)");

    // Export to OBJ format
    vtkNew<vtkOBJExporter> objExporter;
    std::string filePrefix = work_dir + "/out";
    objExporter->SetFilePrefix(filePrefix.c_str());
    objExporter->SetRenderWindow(renderWindow);
    objExporter->Write();

    // Start the interactor
    interactor->Start();

    return 0;
}
