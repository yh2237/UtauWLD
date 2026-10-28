# UtauWLD

WORLD を用いた音声の解析と合成のライブラリです。

- `world/`: WORLD本体（C++）。`LICENSE.txt`などの条件に従います。
- `worldrender/`: Goのラッパー。`refmatch.World`を満たします。
- `speechtiming/`: 発話の時間伸縮モデル。
- `runtimepath/`: ネイティブライブラリの探索。

## ライセンス

MIT License（[LICENSE](LICENSE)）。`world/`以下のC++はWORLD本体とOOuRa FFTの条件に従います（[world/LICENSE.txt](world/LICENSE.txt)）。