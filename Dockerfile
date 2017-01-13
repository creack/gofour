FROM            golang:1.7
MAINTAINER      Guillaume J. Charmes <guillaume@leaf.ag>

# Install linters, coverage tools and test formatters.
RUN             go get github.com/alecthomas/gometalinter && gometalinter -i

# Disable CGO and recompile the stdlib.
ENV             CGO_ENABLED 0
RUN             go install -a -ldflags -d std

# Install jq and yaml2json for parsing glide.lock to precompile.
RUN             apt-get update && apt-get install -y jq
RUN             go get github.com/creack/yaml2json

ARG             APP_DIR=github.com/creack/gofour

ENV             APP_PATH $GOPATH/src/$APP_DIR

WORKDIR         $APP_PATH

# Precompile deps.
# NOTE: From 8 secs to 2 secs to run auth's tests.
ADD             glide.yaml $APP_PATH/glide.yaml
ADD             glide.lock $APP_PATH/glide.lock
ADD             vendor     $APP_PATH/vendor
RUN             yaml2json < glide.lock | \
                jq -r -c '.imports[], .testImports[] | {name, subpackages}' | sed 's/null/[""]/'   | \
                jq -r -c '.name as $name | .subpackages[] | [$name, .] | join("/")' | sed 's|/$||' | \
                while read l; do \
                  echo "$l...";  \
                  go install -ldflags -d $APP_DIR/vendor/$l 2> /dev/null; \
                done

ADD             .          $APP_PATH

RUN             go build -o app -ldflags '-d -w'
