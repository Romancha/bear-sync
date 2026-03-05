import Foundation

/// Protocol abstracting IPC operations for testability.
protocol IPCClientProtocol {
    func getStatus() async throws -> IPCStatusResponse
    func syncNow() async throws -> IPCOkResponse
    func getLogs(lines: Int) async throws -> IPCLogsResponse
    func quit() async throws -> IPCOkResponse
}

extension BridgeIPCClient: IPCClientProtocol {}

/// View model bridging BridgeIPCClient to SwiftUI state.
///
/// Polls the daemon for status at a configurable interval and exposes
/// @Published properties for the menu bar UI.
@MainActor
final class StatusViewModel: ObservableObject {

    @Published var syncStatus: SyncStatus = .idle
    @Published var lastSyncTime: Date?
    @Published var lastError: String?
    @Published var stats: SyncStats = SyncStats()
    @Published var isSyncing: Bool = false
    @Published var bridgeConnected: Bool = false

    private let ipcClient: IPCClientProtocol
    private let pollInterval: TimeInterval
    private var pollTask: Task<Void, Never>?

    var lastSyncDescription: String {
        guard let lastSync = lastSyncTime else {
            return "Never"
        }
        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .full
        return formatter.localizedString(for: lastSync, relativeTo: Date())
    }

    var statusColor: String {
        syncStatus.iconColor
    }

    init(ipcClient: IPCClientProtocol, pollInterval: TimeInterval = 5) {
        self.ipcClient = ipcClient
        self.pollInterval = pollInterval
    }

    /// Start polling the daemon for status updates.
    func startPolling() {
        stopPolling()
        pollTask = Task { [weak self] in
            guard let self else { return }
            while !Task.isCancelled {
                await self.refreshStatus()
                try? await Task.sleep(nanoseconds: UInt64(self.pollInterval * 1_000_000_000))
            }
        }
    }

    /// Stop polling.
    func stopPolling() {
        pollTask?.cancel()
        pollTask = nil
    }

    /// Fetch status once from the daemon.
    func refreshStatus() async {
        do {
            let response = try await ipcClient.getStatus()
            applyStatus(response)
            bridgeConnected = true
        } catch {
            bridgeConnected = false
        }
    }

    /// Trigger an immediate sync via IPC.
    func syncNow() async {
        guard !isSyncing else { return }
        isSyncing = true
        syncStatus = .syncing
        do {
            _ = try await ipcClient.syncNow()
            // After triggering, poll immediately to get updated state
            try? await Task.sleep(nanoseconds: 500_000_000)
            await refreshStatus()
        } catch {
            lastError = error.localizedDescription
            syncStatus = .error
        }
        isSyncing = false
    }

    // MARK: - Private

    private func applyStatus(_ response: IPCStatusResponse) {
        syncStatus = SyncStatus(rawValue: response.state) ?? .idle
        if !response.lastSync.isEmpty, let date = ISO8601DateFormatter().date(from: response.lastSync) {
            lastSyncTime = date
        }
        lastError = response.lastError.isEmpty ? nil : response.lastError
        stats = SyncStats(
            notesCount: response.stats.notesSynced,
            tagsCount: response.stats.tagsSynced,
            queueCount: response.stats.queueProcessed,
            lastDurationMs: Int(response.stats.lastDurationMs)
        )
    }
}
