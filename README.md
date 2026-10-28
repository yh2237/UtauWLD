# UtauWLD

WORLD を用いた音声の解析と合成のライブラリです。

- `world/`: WORLD 本体（C++）。`LICENSE.txt` などの条件に従います。
- `worldrender/`: Go のラッパー。`refmatch.World` を満たします。
- `speechtiming/`: 発話の時間伸縮モデル。
- `runtimepath/`: ネイティブライブラリの探索。

## ビルド

`world/` の C++ をビルドして `utautts-world-engine` を作り、`runtimepath` が探す場所に置いてください。

## ライセンス

MIT License（[LICENSE](LICENSE)）。`world/` 以下の C++ は WORLD 本体と OOuRa FFT の条件に従います（[world/LICENSE.txt](world/LICENSE.txt)）。