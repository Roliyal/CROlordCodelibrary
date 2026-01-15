# ---------- build ----------
FROM debian:bookworm AS build
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates git curl zip unzip tar pkg-config \
    build-essential cmake ninja-build \
  && rm -rf /var/lib/apt/lists/*

RUN git clone --depth=1 https://github.com/microsoft/vcpkg /opt/vcpkg \
 && /opt/vcpkg/bootstrap-vcpkg.sh -disableMetrics

ENV VCPKG_ROOT=/opt/vcpkg
ENV PATH="/opt/vcpkg:${PATH}"

COPY . /src/cpp-svc/
COPY .env /src/.env

RUN /opt/vcpkg/vcpkg install --x-manifest-root=/src/cpp-svc --triplet=x64-linux

RUN cmake -S /src/cpp-svc -B /src/cpp-svc/build \
    -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_TOOLCHAIN_FILE=/opt/vcpkg/scripts/buildsystems/vcpkg.cmake \
 && cmake --build /src/cpp-svc/build

# ---------- runtime ----------
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/cpp-svc/build/cpp-svc /app/cpp-svc
COPY --from=build /src/.env /app/.env
ENV DOTENV_PATH=/app/.env

EXPOSE 8083
CMD ["/app/cpp-svc"]
