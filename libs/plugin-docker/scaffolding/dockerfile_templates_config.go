package scaffolding

// dockerignore keeps what no image needs out of the build context (the
// workspace): history, installed dependencies and smith's own state. dist/
// stays: an image copies the app's build from it.
const dockerignore = `.git
**/node_modules
.smith
`

// goDockerfile builds the binary inside Docker, so the image is Linux
// whatever the developer's OS, from the whole workspace (go.work and the
// modules the app uses), and ships it alone on distroless.
const goDockerfile = `# The app's image (smith run image). The context is the workspace.
FROM golang:[[.GoVersion]]-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app ./[[.Dir]]

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
`

// webDockerfile serves the static build smith writes to dist/<app dir>.
const webDockerfile = `# The app's image (smith run image, after its build). The context is the
# workspace.
FROM nginx:1.29-alpine
COPY dist/[[.Dir]]/ /usr/share/nginx/html/
EXPOSE 80
`
