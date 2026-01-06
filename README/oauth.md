# OAuth

[OAuth. 2.0](https://datatracker.ietf.org/doc/html/rfc6749) is an open standard for **access delegation**, i.e. a way for internet users to authenticate into some webapp, by using an OAuth provider. For example, an app implementing OAuth, may offer a provider such as Google, so that user could authenticate into the app, using a token granted by Google (or other OAuth provider).

## Dropbox Provider

Let's imagine our app implements OAuth, and offers users to be able to authenticate using their Dropbox credentials. The flow goes like this:

1. User clicks on an **OAuth Dropbox** button in our app.
2. User is redirected to Dropbox, where they **authorize** our app to access limited information in their Dropbox account.
3. Once they authorize our app, they're redirected back to our website, with the authorization code from Dropbox.
4. Our **app** reads this code (from the URL query params), and requests an **API token** from the **Dropbox API**.
5. Once the Dropbox API responds with this token, our app can use it to access the user's Dropbox account (until access is revoked).

## OAuth Package

The [oauth2](https://pkg.go.dev/golang.org/x/oauth2) package provides support for making OAuth2 authorized and authenticated HTTP requests. It even includes a [configuration example](https://pkg.go.dev/golang.org/x/oauth2#example-Config-CustomHTTP):

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	conf := &oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		Scopes:       []string{"SCOPE1", "SCOPE2"},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://provider.com/o/oauth2/token",
			AuthURL:  "https://provider.com/o/oauth2/auth",
		},
	}

	// Redirect user to the consent page in the provider.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	_ = client
}
```

Let's go over the `oauth2.Config`:

- `ClientID` is the way we have to identify the OAuth provider. We get it from the provider itself, when we register our app with them.
- `ClientSecret` is the way we have to identify our app with the OAuth provider. We also get it from the provider, when we register our app with them.
- The `Scopes` are the permissions that our app will have in the provider, once the user authorize it.
- The `Endpoint` section includes two URLs:
  1. `AuthURL` in the provider, where we send the user so she can authorize our app to access her data in the provider.
  2. `TokenURL` is where our app will request the **API token** we mentioned at the beginning.

The configuration above is something we'll be using in our code. Regarding the rest of the code:

- The `url` is set after calling the `conf.AuthCodeURL` function. The example above just print it to the terminal, but in our app, that's the URL we'll send the user so she can authorize our app with the provider.
- Once the user authorize our app, we use an `http.Client` to make a request for the **API token**. We'll be using that token each time we want to do something related with the user account in the provider.

## Dropbox as OAuth Provider

We'll be using [Dropbox](https://www.dropbox.com) as our OAuth provider, so we need to create an account with them (which you can do using our Google Account as OAuth provider 😂). Once we have an account, we have to visit [https://www.dropbox.com/developers](https://www.dropbox.com/developers), and **create an app**.

![create](./img/create-dropbox.png)

Then we have to choose the scope access and name for our app:

![create 2](./img/create-dropbox2.png)

And take care of some configuration details:

![create 3](./img/create-dropbox3.png)

> [!WARNING]
> Don't forget to **submit** the changes after setting up the app!

There are several pieces of info in this page, but the relevant ones for getting started are:

- **App key**, which we should use as `ClientID`.
- **App secret**, which we should use as `ClientSecret`.
- **OAuth2 - Redirect URIs**, where Dropbox will redirect the user once she authorizes our app to use her Dropbox account. We have to set this up here, for example, during local development we could use `http://localhost/oauth`.

It's a good idea to add the **app key and secret** in our `.env` and our `.env.prod` files:

```sh
DROPBOX_APP_KEY=<some-key>
DROPBOX_APP_SECRET=<some-secret>
```

Finally, in the **permissions** tab we have to make sure we tick on the `files.content.read`, so that our app can access pics from the user's Dropbox account.

![permissions](./img/permissions-dropbox.png)

## Testing the Setup

First of all, let's install the oauth2 Go Package:

```sh
go get -u golang.org/x/oauth2
```

And adapting the code sample from the beginning to make use of our config shouldn't be difficult; the [documentation](https://www.dropbox.com/developers/documentation) is quite useful, and there's also an [OAuth guide](https://developers.dropbox.com/oauth-guide):

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx := context.Background()

	conf := &oauth2.Config{
		ClientID:     os.Getenv("DROPBOX_APP_KEY"),
		ClientSecret: os.Getenv("DROPBOX_APP_SECRET"),
		Scopes: []string{
			"files.metadata.read",
			"files.content.read",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/oauth2/authorize",
			TokenURL: "https://api.dropboxapi.com/oauth2/token",
		},
	}

	// Redirect user to the consent page in the provider.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	fmt.Printf("Once you have the code, paste it and press enter:\n")

	// Use the authorization code that is pushed to the redirect
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	apiUrl := "https://api.dropboxapi.com/2/files/list_folder"
	jsonPayload := strings.NewReader(`{
		"path": "",
		"recursive": false
	}`)
	resp, err := client.Post(apiUrl, "application/json", jsonPayload)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
```

If you run this code:

```sh
go run cmd/exp/main.go 
```

We should get the output:
```
Visit the URL for the auth dialog: https://www.dropbox.com/oauth2/authorize?access_type=offline&client_id=ypmx1qg2&response_type=code&scope=files.metadata.read+files.content.read&state=state
Once you have the code, paste it and press enter:
```

We have to copy the URL, paste it in our browser where the user is logged in in her Dropbox account, and we'll see a **warning**:

![warning](./img/warning.png)

We click, then the user decides if she wants to **authorize** our app:

![authorize](./img/authorize.png)

If she authorizes our app, we'll get the code:

![code](./img/code.png)

Copy and paste in the terminal to finish the experiment! This is the final output:

```
{"entries": [{".tag": "file", "name": "gopher.jpg", "path_lower": "/gopher.jpg", "path_display": "/gopher.jpg", "id": "id:QvDW163Du64AAAAAABg", "client_modified": "2026-01-06T14:53:41Z", "server_modified": "2026-01-06T14:53:41Z", "rev": "01647b9576f960c000000032621fc63", "size": 44732, "is_downloadable": true, "content_hash": "0737c7175e1f154c7a82f8320788273baa"}], "cursor": "AmlmcFehbrWL466u7UywwbQ8Ky9QvhzKDeCospPr0", "has_more": false}
```