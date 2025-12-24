package docs

// CustomSwaggerHTML is the custom HTML template for the Swagger UI.
// It overrides the default interface to inject a custom script that automatically
// captures the access token from the login response and applies it to the "Authorize" button.
//
// Key Features:
// - Custom Styles: Applies local stylesheets.
// - Auto-Authorization: Intercepts the response from /api/tuya/auth, extracts the access_token,
//   and programmatically triggers the Swagger UI authorization action with "Bearer <token>".
const CustomSwaggerHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
    <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
    <style>
      html {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after {
        box-sizing: inherit;
      }

      body {
        margin: 0;
        background: #fafafa;
      }
    </style>
  </head>

  <body>
    <div id="swagger-ui"></div>

    <script src="./swagger-ui-bundle.js"></script>
    <script src="./swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function () {
        // Build a system
        const ui = SwaggerUIBundle({
          url: "doc.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          defaultModelsExpandDepth: -1,
          defaultModelExpandDepth: 3,
          displayRequestDuration: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout",
          responseInterceptor: (response) => {
            // Check if this is the auth endpoint
            if (response.url && response.url.indexOf("/api/tuya/auth") > -1 && response.status === 200) {
                try {
                    console.log("Login detected, attempting to extract token...");
                    // Parse body if it isn't an object already
                    let body = response.body; 
                    if (typeof body === 'string') {
                        try {
                            body = JSON.parse(body);
                        } catch(e) {}
                    }
                    // Often response.obj is already populated by Swagger
                    const data = (body && body.data) || (response.obj && response.obj.data);

                    if (data && data.access_token) {
                        const token = data.access_token;
                        console.log("Token found:", token);
                        
                        // The security definition name in main.go is "BearerAuth"
                        const securityDefinition = "BearerAuth";
                        const bearerToken = "Bearer " + token;

                        // Trigger the authorization action
                        ui.authActions.authorize({
                            [securityDefinition]: {
                                name: securityDefinition,
                                schema: {
                                    type: "apiKey",
                                    in: "header",
                                    name: "Authorization",
                                    description: "Type 'Bearer' followed by a space and JWT token."
                                },
                                value: bearerToken
                            }
                        });
                        console.log("Token applied to Swagger UI!");
                    }
                } catch (e) {
                    console.error("Error auto-filling token:", e);
                }
            }
            return response;
          }
        });

        window.ui = ui;
      };
    </script>
  </body>
</html>
`