---
title: go-bundler - AtCoder 向け Go ソースバンドラー
description: 競技プログラミングと AtCoder の単一ファイル提出向けの軽量な Go ソースバンドラー。
---

# go-bundler

🇬🇧 English: [../](../)

**go-bundler** は競技プログラミング向けの軽量な Go ソースバンドラーです。

複数の Go ソースファイルを 1 つの Go ファイルにマージするので、
AtCoder、yukicoder など、単一ファイル提出が必要なオンラインジャッジに
Go プログラムを提出しやすくなります。

## 機能

- Go パッケージを 1 つのソースファイルにバンドル
- RTA (Rapid Type Analysis) によるデッドコード除去
- ジェネリクス、埋め込み構造体、インターフェイス型に対応
- 単一コマンドで標準出力またはファイルに出力
- ファイル変更を検知して再ビルドする watch モード
- 行数メトリクスおよびサステナビリティメトリクスの任意出力

## インストール

```sh
go install github.com/Atnuhs/go-bundler@latest
```

## 使い方

```sh
go-bundler -dir ./path/to/your/package > submit.go
```

### オプション

| フラグ | 説明 |
|---|---|
| `-dir` | 対象パッケージのディレクトリ (デフォルト: `.`) |
| `-o`, `-out` | 出力ファイルパス (デフォルト: 標準出力) |
| `-watch` | ローカルパッケージファイルを監視して変更時に再ビルドする (`-o` が必要) |
| `-with-metrics` | 行数メトリクスをコメントブロックとして出力する |
| `-with-sustainability-metrics` | CO2 と樹木換算のメトリクスを出力する |

## 使用例

パッケージをバンドルしてファイルに書き出す:

```sh
go-bundler -dir ./cmd/abc123 > submit.go
```

行数メトリクス付きでバンドル:

```sh
go-bundler -dir ./cmd/abc123 -with-metrics > submit.go
```

標準出力をリダイレクトする代わりに、直接ファイルへ書き出す:

```sh
go-bundler -dir ./cmd/abc123 -o submit.go
```

watch モードでは、解決されたローカルパッケージ配下の `.go` ファイルが変更されるたびに `submit.go` を再ビルドします (Ctrl+C で終了):

```sh
go-bundler -dir ./cmd/abc123 -o submit.go -watch
```

## リンク

- [GitHub リポジトリ](https://github.com/Atnuhs/go-bundler)
- [Qiita: Go のバンドラーを書いた](https://qiita.com/Authns/items/ddba6d392ec6a316383f)
