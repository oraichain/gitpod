# Copyright (c) 2020 Gitpod GmbH. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

FROM alpine:3.16 as compress

RUN apk add brotli gzip

COPY ./build /www

WORKDIR /www

RUN find . -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.png' -o -name '*.svg' -o -name '*.map' -o -name '*.json' \) \
    -exec /bin/sh -c 'gzip -v -f -9 -k "$1"' /bin/sh {} \;

RUN find . -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.png' -o -name '*.svg' -o -name '*.map' -o -name '*.json' \) \
    -exec /bin/sh -c 'brotli -v -q 11 -o "$1.br" "$1"' /bin/sh {} \;

FROM eu.gcr.io/gitpod-core-dev/build/dashboard:commit-422ab8bf211f4ce04f386656b2609a9dbfbc0b26

COPY --from=compress /www /www

ARG __GIT_COMMIT
ARG VERSION

ENV GITPOD_BUILD_GIT_COMMIT=422ab8bf211f4ce04f386656b2609a9dbfbc0b26
ENV GITPOD_BUILD_VERSION=lau-time-view-11526.15

# curretn image: eu.gcr.io/gitpod-core-dev/build/dashboard:commit-422ab8bf211f4ce04f386656b2609a9dbfbc0b26
