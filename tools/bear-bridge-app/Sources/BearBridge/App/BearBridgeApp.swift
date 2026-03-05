import SwiftUI

@main
struct BearBridgeApp: App {
    @StateObject private var viewModel: StatusViewModel

    init() {
        let ipcClient = BridgeIPCClient()
        _viewModel = StateObject(wrappedValue: StatusViewModel(ipcClient: ipcClient))
    }

    var body: some Scene {
        MenuBarExtra {
            MenuBarView(viewModel: viewModel)
                .onAppear {
                    viewModel.startPolling()
                }
                .onDisappear {
                    viewModel.stopPolling()
                }
        } label: {
            Image(systemName: menuBarIcon)
                .symbolRenderingMode(.palette)
                .foregroundStyle(menuBarIconColor)
        }
        .menuBarExtraStyle(.window)
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
