import SwiftUI

extension Notification.Name {
    static let openLogViewer = Notification.Name("openLogViewer")
}

@main
struct BearBridgeApp: App {
    @StateObject private var viewModel: StatusViewModel
    @StateObject private var logViewModel: LogViewModel
    @StateObject private var settingsManager: SettingsManager
    @Environment(\.openWindow) private var openWindow
    private let notificationService: NotificationService

    init() {
        let ipcClient = BridgeIPCClient()
        let settings = SettingsManager()
        let notifications = NotificationService()
        notifications.isEnabled = settings.notificationsEnabled
        notifications.onOpenLogViewer = {
            NotificationCenter.default.post(name: .openLogViewer, object: nil)
        }

        _viewModel = StateObject(wrappedValue: StatusViewModel(
            ipcClient: ipcClient,
            notificationService: notifications
        ))
        _logViewModel = StateObject(wrappedValue: LogViewModel(ipcClient: ipcClient))
        _settingsManager = StateObject(wrappedValue: settings)
        self.notificationService = notifications
    }

    var body: some Scene {
        MenuBarExtra {
            MenuBarView(viewModel: viewModel, logViewModel: logViewModel)
                .onAppear {
                    viewModel.startPolling()
                }
                .onDisappear {
                    viewModel.stopPolling()
                }
                .onReceive(NotificationCenter.default.publisher(for: .openLogViewer)) { _ in
                    openWindow(id: "log-viewer")
                }
        } label: {
            Image(systemName: menuBarIcon)
                .symbolRenderingMode(.palette)
                .foregroundStyle(menuBarIconColor)
        }
        .menuBarExtraStyle(.window)

        Window("Bear Bridge Logs", id: "log-viewer") {
            LogViewerWindow(viewModel: logViewModel)
        }
        .defaultSize(width: 700, height: 500)

        Window("Bear Bridge Settings", id: "settings") {
            SettingsWindow(settings: settingsManager)
                .onReceive(settingsManager.$notificationsEnabled) { enabled in
                    notificationService.isEnabled = enabled
                }
        }
        .defaultSize(width: 450, height: 300)
    }

    private var menuBarIcon: String {
        "arrow.triangle.2.circlepath"
    }

    private var menuBarIconColor: Color {
        switch viewModel.syncStatus {
        case .idle: return .primary
        case .syncing: return .yellow
        case .error: return .red
        }
    }
}
