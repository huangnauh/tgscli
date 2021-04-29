# tgscli

CLI for Telegram Files, written in Go.

## Install
Install from source:

```
go get -u github.com/huangnauh/tgscli/cmd/tgscli
```

## Usage

### put
```
❯ tgscli put swe_at_google.pdf
9.16 MiB / 9.16 MiB [-----------------------------] 100.00% 1.40 MiB p/s 6.5s
Upload time:6.542857s
```

### share
```
❯ tgscli share swe_at_google.pdf
```

### get

```
❯ tgscli get swe_at_google.pdf
9.16 MiB / 9.16 MiB [-----------------------------] 100.00% 1.78 MiB p/s 5.4s
Download File to swe_at_google.pdf
```

### list
```
❯ tgscli list

  TIME                           LENGTH  NAME
-------------------------------------------------------------------
  Mon, 01 Jan 0001 00:00:00 UTC  3.2 MB  lusp_fileio_slides.pdf
  Mon, 26 Apr 2021 22:57:04 CST  9.6 MB  swe_at_google.pdf
```

## Read more

[documentation](docs/tgscli.md).