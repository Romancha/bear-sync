import AppKit
import Foundation

// MARK: - Argument Parsing

struct Config {
    let url: String
    let timeout: TimeInterval
}

func printUsage() {
    let usage = """
        Usage: bear-xcall -url <bear://x-callback-url/...> [-timeout <seconds>]

        Options:
          -url       Bear x-callback-url to execute (required)
          -timeout   Timeout in seconds (default: 10)
          --help     Show this help message
        """
    FileHandle.standardError.write(Data(usage.utf8))
}

func parseArgs() -> Config? {
    let args = CommandLine.arguments

    if args.contains("--help") || args.count == 1 {
        printUsage()
        return nil
    }

    var urlValue: String?
    var timeoutValue: TimeInterval = 10

    var i = 1
    while i < args.count {
        switch args[i] {
        case "-url":
            i += 1
            guard i < args.count else {
                FileHandle.standardError.write(Data("Error: -url requires a value\n".utf8))
                return nil
            }
            urlValue = args[i]
        case "-timeout":
            i += 1
            guard i < args.count, let t = TimeInterval(args[i]), t > 0 else {
                FileHandle.standardError.write(Data("Error: -timeout requires a positive number\n".utf8))
                return nil
            }
            timeoutValue = t
        default:
            FileHandle.standardError.write(Data("Error: unknown argument '\(args[i])'\n".utf8))
            printUsage()
            return nil
        }
        i += 1
    }

    guard let url = urlValue else {
        FileHandle.standardError.write(Data("Error: -url is required\n".utf8))
        printUsage()
        return nil
    }

    guard url.hasPrefix("bear://") else {
        FileHandle.standardError.write(Data("Error: URL must start with bear://\n".utf8))
        return nil
    }

    return Config(url: url, timeout: timeoutValue)
}

// MARK: - URL Callback Handler

class XCallbackHandler: NSObject {
    private let timeout: TimeInterval
    private var timeoutTimer: Timer?

    init(timeout: TimeInterval) {
        self.timeout = timeout
        super.init()
    }

    func start(with bearURL: String) {
        // Register URL scheme handler for bear-xcall:// callbacks.
        NSAppleEventManager.shared().setEventHandler(
            self,
            andSelector: #selector(handleURL(_:withReply:)),
            forEventClass: AEEventClass(kInternetEventClass),
            andEventID: AEEventID(kAEGetURL)
        )

        // Inject x-success and x-error callback URLs into the bear:// URL.
        let separator = bearURL.contains("?") ? "&" : "?"
        let callbackURL = bearURL
            + separator
            + "x-success=bear-xcall://x-callback-url/success"
            + "&x-error=bear-xcall://x-callback-url/error"

        guard let url = URL(string: callbackURL) else {
            writeError("Invalid URL: \(bearURL)")
            exit(1)
        }

        NSWorkspace.shared.open(url)

        // Start timeout timer. Write timeout error to stdout (same format as Bear error responses)
        // so the Go caller can parse the structured error via parseXcallResult.
        timeoutTimer = Timer.scheduledTimer(withTimeInterval: timeout, repeats: false) { _ in
            let errorJSON: [String: Any] = [
                "errorCode": -1,
                "errorMessage": "Timeout after \(Int(self.timeout)) seconds waiting for Bear callback",
            ]
            if let data = try? JSONSerialization.data(withJSONObject: errorJSON),
               let str = String(data: data, encoding: .utf8)
            {
                print(str)
            }
            exit(2)
        }
    }

    @objc func handleURL(_ event: NSAppleEventDescriptor, withReply _: NSAppleEventDescriptor) {
        timeoutTimer?.invalidate()

        guard let urlString = event.paramDescriptor(forKeyword: AEKeyword(keyDirectObject))?.stringValue,
              let url = URL(string: urlString)
        else {
            writeError("Failed to parse callback URL")
            exit(1)
        }

        let isError = url.path == "/error"

        // Parse query parameters into a dictionary.
        var result: [String: String] = [:]
        if let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
           let queryItems = components.queryItems
        {
            for item in queryItems {
                if let value = item.value {
                    result[item.name] = value
                }
            }
        }

        // Convert errorCode to integer if present.
        var jsonObject: [String: Any] = result
        if let errorCodeStr = result["errorCode"], let errorCode = Int(errorCodeStr) {
            jsonObject["errorCode"] = errorCode
        }

        // Write all responses to stdout so the Go caller can parse structured
        // error details via parseXcallResult (drop-in replacement for xcall).
        do {
            let data = try JSONSerialization.data(
                withJSONObject: jsonObject,
                options: [.sortedKeys]
            )
            if let str = String(data: data, encoding: .utf8) {
                print(str)
                exit(isError ? 1 : 0)
            } else {
                writeError("Failed to encode response as UTF-8")
                exit(1)
            }
        } catch {
            writeError("Failed to serialize response: \(error)")
            exit(1)
        }
    }

    private func writeError(_ message: String) {
        let errorJSON: [String: Any] = [
            "errorCode": -1,
            "errorMessage": message,
        ]
        if let data = try? JSONSerialization.data(withJSONObject: errorJSON),
           let str = String(data: data, encoding: .utf8)
        {
            FileHandle.standardError.write(Data((str + "\n").utf8))
        }
    }
}

// MARK: - Main

guard let config = parseArgs() else {
    exit(1)
}

let app = NSApplication.shared
app.setActivationPolicy(.accessory)

let handler = XCallbackHandler(timeout: config.timeout)
handler.start(with: config.url)

app.run()
