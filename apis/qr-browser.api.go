package apis

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const index = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>QR Code Scanner For</title>
  </head>
  <body>
    <h1>Scan The QR Code to Log into whatsapp web account</h1>
    <img height="512" width="512" />
    <script>
      // get host url for api calling
      const id = "%s";
      // const url = window.location.href
      const url = ` + "`${window.location.protocol}//${window.location.host}/v1/qr_code/`;" +
	`      function hexToBase64(str) {
    return btoa(
      String.fromCharCode.apply(
        null,
        str
          .replace(/\r|\n/g, '')
          .replace(/([\da-fA-F]{2}) ?/g, '0x$1 ')
          .replace(/ +$/, '')
          .split(' ')
      )
    );
  }
  async function CallAPIAndSetImage() {
    const res = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'X-Api-Token': id,
      },
    });
    try {
      const data = await res.json();
      if(data.qrCode){
        const img = document.querySelector('img');
        img.src = 'data:image/png;base64,' + data.qrCode;
      }
    } catch (error) {}
  }
  CallAPIAndSetImage();
  setInterval(() => CallAPIAndSetImage(), 5000);
      // const img = document.querySelector('img')
      // const url = window.location.href
    </script>
  </body>
</html>`

func OpenBrowserWithQr(c *fiber.Ctx) error {
	id := c.Params("id")
	// println(c.Hostname())
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send([]byte(fmt.Sprintf(index, id)))
}
