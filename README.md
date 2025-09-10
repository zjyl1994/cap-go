# cap-go

**cap-go** is a Go-based backend framework that provides full compatibility with the [Cap.js](https://capjs.js.org/) implementation.

## Installation

Install the package using Go's module system:

```bash
go get github.com/zjyl1994/cap-go
```

## Usage

To integrate `cap-go` into your application, follow these steps:

1. Register the `/cap/challenge` route and connect it to the `CreateChallenge` method.
2. Register the `/cap/redeem` route and connect it to the `RedeemChallenge` method.

On the frontend, use the `@cap.js/widget` component and set the `data-cap-api-endpoint` attribute to point to your API base path:

```html
<cap-widget id="cap" data-cap-api-endpoint="/cap/"></cap-widget>
<script>
  document.getElementById('cap').addEventListener("solve", (e) => {
    const token = e.detail.token;
    console.log(token); // Handle the token as needed
  });
</script>
```

Once verification completes on the frontend, you can validate the received token on the server using the `ValidateToken` method to ensure its authenticity.

For a complete, runnable example, see [`examples/main.go`](examples/main.go).

## Documentation

- [Official Cap.js Documentation](https://capjs.js.org/)

## License

This project is open source and licensed under the MIT License. 
You are free to use, modify, and distribute the code in accordance with the terms of the license.
