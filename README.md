<div align="center">
	<img width="500" src=".github/logo.svg" alt="pinpt-logo">
</div>

<p align="center" color="#6a737d">
	<strong>HTTPClient is a zero dependency client which supports pluggable retry and pagination</strong>
</p>

## Setup

	go get -u github.com/pinpt/httpclient

## Usage

The most basic usage which mimics http.Default:

```golang
httpclient.Default.Get("https://foo.com")
```

Customize the client by setting Config:

```golang
config := httpclient.NewConfig()
config.Retryable = httpclient.NewBackoffRetry(time.Millisecond, time.Millisecond*50, 10*time.Second, 2)
client := httpclient.NewHTTPClient(context.Background(), config, http.DefaultClient)
resp, err := client.Get("https://foo.com")
```

## Pluggable

The httpclient package is very customizable.  You can pass in any implementation of the Client interface which `http.Client` implements.  You can implement the Retryable and Paginator interfaces for customizing how to Retry failed requests and how to handle pagination.

## License

Copyright &copy; 2018 by PinPT, Inc. MIT License. Pull requests welcome! 🙏🏻
