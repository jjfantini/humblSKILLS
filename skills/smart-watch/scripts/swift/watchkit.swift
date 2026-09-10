// watchkit — native perception helper for the smart-watch skill.
//
//   watchkit ocr <image...>            TSV: <path>\t<text lines joined by " | ">
//   watchkit transcribe <wav> [locale] TSV: <start>\t<end>\t<text>  (seconds)
//
// Built by scripts/lib.sh::ensure_watchkit with `swiftc -O -parse-as-library`.
// OCR uses Vision (macOS 10.15+). Transcription uses SpeechAnalyzer, which
// only exists in the macOS 26 SDK; the `compiler(>=6.2)` guard stands in for
// "SDK 26+" so the binary still builds on older toolchains with OCR only.

import Foundation
import AppKit
import Vision
#if compiler(>=6.2)
import Speech
import AVFoundation
#endif

@main struct WatchKit {
    static func main() async {
        let args = Array(CommandLine.arguments.dropFirst())
        guard let cmd = args.first else { usage() }
        switch cmd {
        case "ocr": ocr(Array(args.dropFirst()))
        case "transcribe": await transcribe(Array(args.dropFirst()))
        default: usage()
        }
    }

    static func usage() -> Never {
        FileHandle.standardError.write("usage: watchkit ocr <image...> | watchkit transcribe <wav> [locale]\n".data(using: .utf8)!)
        exit(2)
    }

    // MARK: ocr

    static func ocr(_ paths: [String]) {
        guard !paths.isEmpty else { usage() }
        var out = [String](repeating: "", count: paths.count)
        let lock = NSLock()
        DispatchQueue.concurrentPerform(iterations: paths.count) { i in
            let text = recognize(path: paths[i])
            lock.lock(); out[i] = text; lock.unlock()
        }
        var buf = ""
        for (i, p) in paths.enumerated() { buf += "\(p)\t\(out[i])\n" }
        FileHandle.standardOutput.write(buf.data(using: .utf8)!)
    }

    static func recognize(path: String) -> String {
        guard let img = NSImage(contentsOf: URL(fileURLWithPath: path)) else { return "" }
        var rect = NSRect(origin: .zero, size: img.size)
        guard let cg = img.cgImage(forProposedRect: &rect, context: nil, hints: nil) else { return "" }
        let req = VNRecognizeTextRequest()
        req.recognitionLevel = .accurate
        req.usesLanguageCorrection = true
        do { try VNImageRequestHandler(cgImage: cg).perform([req]) } catch { return "" }
        let lines = (req.results ?? []).compactMap { obs -> String? in
            guard let c = obs.topCandidates(1).first, c.confidence >= 0.3 else { return nil }
            let s = c.string.replacingOccurrences(of: "\t", with: " ")
                              .replacingOccurrences(of: "\n", with: " ")
                              .trimmingCharacters(in: .whitespaces)
            return s.isEmpty ? nil : s
        }
        return lines.joined(separator: " | ")
    }

    // MARK: transcribe

    static func transcribe(_ args: [String]) async {
        guard let path = args.first else { usage() }
        #if compiler(>=6.2)
        if #available(macOS 26, *) {
            let localeID = args.count > 1 ? args[1] : "en_US"
            do { try await runSpeechAnalyzer(path: path, localeID: localeID) }
            catch {
                FileHandle.standardError.write("transcribe failed: \(error)\n".data(using: .utf8)!)
                exit(1)
            }
            return
        }
        #endif
        FileHandle.standardError.write("transcribe: SpeechAnalyzer needs macOS 26+; use the whisper-cpp fallback\n".data(using: .utf8)!)
        exit(3)
    }

    #if compiler(>=6.2)
    @available(macOS 26, *)
    static func runSpeechAnalyzer(path: String, localeID: String) async throws {
        let transcriber = SpeechTranscriber(locale: Locale(identifier: localeID),
                                            transcriptionOptions: [],
                                            reportingOptions: [],
                                            attributeOptions: [.audioTimeRange])
        if let req = try await AssetInventory.assetInstallationRequest(supporting: [transcriber]) {
            FileHandle.standardError.write("checking speech assets for \(localeID) (downloads once per machine)...\n".data(using: .utf8)!)
            try await req.downloadAndInstall()
        }
        let analyzer = SpeechAnalyzer(modules: [transcriber])
        let file = try AVAudioFile(forReading: URL(fileURLWithPath: path))

        // Span = first run start -> last run end. `runs.first?.audioTimeRange`
        // alone covers only the first word (measured 0.24s for a 2.2s sentence).
        async let drain: Void = {
            for try await r in transcriber.results {
                let ranges = r.text.runs.compactMap { $0.audioTimeRange }
                let start = ranges.first?.start.seconds ?? -1
                let end = ranges.last?.end.seconds ?? -1
                let text = String(r.text.characters)
                    .replacingOccurrences(of: "\t", with: " ")
                    .replacingOccurrences(of: "\n", with: " ")
                    .trimmingCharacters(in: .whitespaces)
                if text.isEmpty { continue }
                let line = String(format: "%.2f\t%.2f\t%@\n", start, end, text)
                FileHandle.standardOutput.write(line.data(using: .utf8)!)
            }
        }()
        if let last = try await analyzer.analyzeSequence(from: file) {
            try await analyzer.finalizeAndFinish(through: last)
        } else {
            await analyzer.cancelAndFinishNow()
        }
        try await drain
    }
    #endif
}
