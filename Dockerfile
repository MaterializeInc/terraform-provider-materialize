FROM golang:1.24-alpine

COPY --from=hashicorp/terraform:1.11.4 /bin/terraform /bin/terraform

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o ~/.terraform.d/plugins/materialize.com/devex/materialize/0.1/linux_amd64/terraform-provider-materialize

WORKDIR /usr/src/app/integration
