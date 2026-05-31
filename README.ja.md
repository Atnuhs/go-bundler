# go-bundler

🇬🇧 English: [README.md](README.md)

ドキュメント: <https://atnuhs.github.io/go-bundler/ja/>
解説記事: <https://qiita.com/Authns/items/ddba6d392ec6a316383f>

go-bundler は競技プログラミング向けの軽量な Go ソースバンドラーです。
ローカルの Go パッケージファイルを再帰的に解決し、到達不能なコードを削除（ツリーシェイキング）して、
1 つのバンドル済みファイルを出力します。主に AtCoder、yukicoder など、
単一ファイル提出が必要なオンラインジャッジで使うことを想定しています。

## 機能

- RTA (Rapid Type Analysis) によるデッドコード除去
- ジェネリクス、埋め込み構造体、インターフェイス型に対応
- 単一コマンドで標準出力またはファイルに出力
- ファイル変更を検知して再ビルドする watch モード
- 行数メトリクスおよびサステナビリティメトリクスの任意出力

## インストール

```bash
go install github.com/Atnuhs/go-bundler@latest
```

## 使い方

```bash
go-bundler -dir ./path/to/your/package > bundled.go
```

`go-bundler` は Go パッケージを 1 つのソースファイルにバンドルします。デフォルトではバンドル済みコードのみを標準出力に出力します。

`-o`/`--out` でファイルに直接書き出したり、`-watch` で変更検知時に自動再ビルドできます:

```text
  -dir string
        対象パッケージのディレクトリ (デフォルト ".")
  -o string
        出力ファイルパス (-out のショートハンド)
  -out string
        出力ファイルパス (デフォルト: 標準出力)
  -watch
        ローカルパッケージファイルを監視して変更時に再ビルドする (-o が必要)
  -with-metrics
        go-bundler のメトリクスコメントブロックを出力する
  -with-sustainability-metrics
        サステナビリティメトリクス (CO2、樹木換算) のコメントブロックを出力する
```

## 使用例

シンプルにバンドルファイルを出力:

```bash
go-bundler -dir ./cmd/app > bundled.go
```

行数メトリクス付きでバンドルファイルを出力:

```bash
go-bundler -dir ./cmd/app -with-metrics > bundled.go
```

行数メトリクスとサステナビリティメトリクス付きでバンドルファイルを出力:

```bash
go-bundler -dir ./cmd/app -with-metrics -with-sustainability-metrics > bundled.go
```

標準出力をリダイレクトする代わりに、直接ファイルへ書き出す:

```bash
go-bundler -dir ./cmd/app -o bundled.go
```

watch モードでは、解決されたローカルパッケージ配下の `.go` ファイルが変更されるたびに `bundled.go` を再ビルドします (Ctrl+C で終了):

```bash
go-bundler -dir ./cmd/app -o bundled.go -watch
```

`-with-sustainability-metrics` を有効にすると、`go-bundler` はモデルベースでざっくり推定した CO2 削減量と、

それに相当する樹木の本数を含む追加のメトリクスブロックを出力します。

これらの値はあくまで参考値であり、実際の環境負荷を示すものではありません。

## ライセンス

MIT License
