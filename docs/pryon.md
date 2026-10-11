# Native Amazon wake-word support

EchoLocal can use the Pryon detector already present in the Fire OS 5 firmware of the
second-generation Echo Dot. This adds **Alexa** to the wake-word select exposed through ESPHome and
Home Assistant. The detector is wake-only: EchoLocal continues to own audio capture after the wake,
Home Assistant Assist, LEDs, TTS, media, ducking and Sendspin.

## Requirements

- an Echo Dot 2 (`biscuit`) on **Fire OS 5** (Android SDK 22), already running EchoLocal with
  root and SELinux permissive;
- the original firmware files that supply Amazon's SpeechInteractionManager, Pryon libraries and
  locale models;
- local Wi-Fi for ESPHome/Home Assistant discovery;
- Go, ADB, Java and an Android SDK to build from the source tree; Windows PowerShell 5.1 or later
  for the provisioning script.

Amazon account registration is not required after the device is unlocked. EchoLocal does not ship,
copy off the device or redistribute Amazon binaries or model files.

### Why Fire OS 5

The main `echolocal install` targets Fire OS 6, where biscuit has no Android framework. Pryon needs
one: the companion is a privileged APK that Android's package manager has to scan, and the media
bridge shares Pryon's capture through `AudioRecord`. So Pryon is installed separately, onto a Dot
that already runs echod on Fire OS 5. echod itself still supports Fire OS 5 through the shared
arm64 build.

The base install onto a fresh Fire OS 5 Dot — boot image, ledcontroller takeover, name and key — is
not part of this tree. Do it once with the original fork's installer (branch
`agent/pryon-alexa-wakeword`: `.\provision-echo-dot.ps1 -Name "Kitchen Echo"`) or echoctl 0.0.6.

## Installing

On Windows, with the Dot connected by USB:

```powershell
.\provision-echo-dot.ps1
```

Use `-Serial G090XXXXXXXXXXXX` when more than one device is attached. The script:

1. selects only the requested `biscuit` device and verifies SDK 22, root, permissive SELinux and an
   installed echod;
2. builds the arm64 echod, the echolocal CLI, the Android media bridge and the Pryon companion;
3. saves a gitignored rollback snapshot;
4. runs `echolocal pryon install`, which discovers the required Amazon files on the Dot, installs
   echod, the bridge and the companion, reboots when Android must scan the companion, and restores
   `/system` read-only;
5. verifies the ESPHome API, mDNS identity, Android media bridge, Pryon frame counter and shared
   live microphone path; and
6. saves a private Home Assistant credential receipt in the gitignored rollback directory and
   prints the device's ESPHome encryption key last.

`-SkipBuild` reuses already built outputs. `-BuildOnly` builds without touching a device.

Elsewhere, build the pieces and run the echolocal CLI directly:

```sh
make build-echod DOT_ARCH=arm64      # Fire OS 5 runs an arm64 kernel
make build-echolocal                 # the host CLI, bin/echolocal
# android/pryon/build.ps1 and android/amazon-helper/build.ps1 build the APK and the JAR
bin/echolocal pryon install --echod bin/echod-arm64 --reboot
```

`--apk` and `--helper` default to the build outputs under `android/`. Re-running is safe: a run with
nothing new skips the reboot and only re-verifies.

## Home Assistant

Add the discovered EchoLocal ESPHome device and paste the encryption key printed by the installer.
The same key is saved as `home-assistant-credentials.txt` inside the rollback snapshot reported by
the script, so it can be recovered after the terminal closes. Keep that file private.
Choose **Alexa** in the desired assistant's Wake word select. Bundled microWakeWord choices and
downloaded openWakeWord/microWakeWord models remain available in other assistant slots.

Only **Alexa** is exposed by this integration. Some Amazon firmware images also contain native
models for `Amazon`, `Computer` or `Echo`, but changing Pryon models requires restarting its Android
process. Those words should be added as a separate, tested change rather than inferred from files
that may differ by firmware and locale.

### Firmware updates

Home Assistant's update entity follows upstream's release channel, whose builds carry no Android
media bridge. On a Pryon device echod refuses that install, and the card reports the failure. Update
with `echolocal pryon install --echod` (or the provisioning script) instead.

## Architecture and security boundary

The privileged `com.echolocal.pryon` APK initializes Amazon's on-device detector and sends bounded
wake metadata to EchoLocal. Pryon's native service owns the single privileged `AudioRecord` and its
Amazon `AudioStream`; EchoLocal's Android media helper reads conversation PCM through a second
authenticated reader on that same stream. This avoids the Fire OS single-input race caused by two
independent recorders. The helper continues to own `AudioTrack` playback. EchoLocal authenticates
the companion process by Android UID before accepting wake events. No audio is sent to Amazon or
another remote service by this integration.

Playback goes through `AudioTrack`, and Fire OS's audio service applies Amazon's speaker tuning (the
`EQ_*.cfg` FIR and `MBCL.cfg` compressor in `/vendor/etc/audio-algorithms`) to it. echod therefore
leaves its own copy of that tuning out, and the **Speaker EQ** switch stays off: running both
applies the tuning twice, which sounds muffled and hissy.

In echod, the bridge's audio joins the microphone path at leveling: there is no array to beamform
and no loopback to cancel against, so those stages are skipped. Its byte stream is cut into the
20 ms frames everything downstream counts time in.

## Verification and rollback

`echolocal pryon status` reports the ESPHome API, Android media bridge and Pryon configuration.
A successful install requires `PRYON_CAPTURE_ACTIVE`, `PRYON_SHARED_PCM_FIRST_FRAME` and
`PRYON_READY`; speaking “Alexa” should then produce a wake event and start the selected Home
Assistant Assist pipeline. The cyan startup walk waits up to 60 seconds for Home Assistant to
subscribe to a voice pipeline, then fades out rather than looking like a recovery or boot loop.

The companion owns `/system/priv-app/EchoLocalPryon`, while the media helper and Pryon UID marker are
stored under `/data/misc/echolocal`. `echolocal pryon remove` force-stops the companion, removes
only those EchoLocal-owned files and puts echod back on direct ALSA; reboot afterwards so the
package goes.
To restore an earlier echod, use the binary saved in the provisioning script's snapshot. Never remove
or replace Amazon's SpeechInteractionManager, native libraries or model directories.

## Keeping up with upstream

This fork is kept as a short series of commits on top of
[ygelfand/echolocal](https://github.com/ygelfand/echolocal), and almost everything it adds lives in
files upstream does not have:

| Area | Fork-owned files | Hooks in upstream files |
| --- | --- | --- |
| Android media bridge | `internal/android/amazon/`, `internal/hardware/{mic,speaker}/android.go`, `internal/layout/pryon.go` | `mic.go` (Start/Run/Close, and `deliver` split out of `broadcast`), `acquire.go` (Start), `speaker.go` (Run/Close, `available`, and `New` skipping the driver tuning), `startup.go` in both (startup progress) |
| Alexa routing | `internal/feature/pryon/`, `internal/lib/wake/pryon.go`, `internal/feature/detect/pryon.go`, `internal/component/all/all_pryon.go` | `lib/wake/wake.go` (`Lib` returns the Pryon-aware library), `feature/detect/detect.go` (one call: `choose` keeps Pryon out of the PCM engine), `feature/voice/conversation.go` (one call: Alexa as first-run default) |
| Startup ring | — | `hardware/led/splash.go` (startup walk gives up after 60 s) |
| Installer | `internal/host/installer/pryon.go`, `internal/cli/echolocal/pryon.go`, `provision-echo-dot.ps1` | one line in `echolocal/root.go` |
| Update guard | `internal/update/pryon.go` | one check in `update/install.go` |

Upstream keeps its wake-word library, detector and voice turn in a separate module,
[`libcountertop`](https://github.com/ygelfand/libcountertop). The fork does not patch it: Pryon is
wrapped around it in the files above, so a change there shows up as a build error in
`internal/lib/wake/pryon.go` or `internal/feature/detect/pryon.go` rather than as a conflict.

To take upstream's changes:

```sh
git config rerere.enabled true   # once: reuse earlier conflict resolutions
git fetch upstream
git rebase upstream/main
go build ./... && go test ./...
```

Conflicts, when there are any, are confined to the hooks above.
