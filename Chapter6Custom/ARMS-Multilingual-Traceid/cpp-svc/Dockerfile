FROM debian:bookworm AS build

WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates git curl zip unzip tar pkg-config \
    build-essential ninja-build cmake \
  && rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/microsoft/vcpkg /opt/vcpkg \
 && cd /opt/vcpkg \
 && git checkout 2024.12.16 \
 && /opt/vcpkg/bootstrap-vcpkg.sh -disableMetrics

COPY vcpkg.json /src/vcpkg.json
COPY CMakeLists.txt /src/CMakeLists.txt

# 安装依赖（BuildKit cache）
RUN --mount=type=cache,target=/root/.cache/vcpkg \
    --mount=type=cache,target=/opt/vcpkg/downloads \
    /opt/vcpkg/vcpkg install --x-manifest-root=/src --triplet=x64-linux

COPY main.cpp /src/main.cpp
COPY .env /src/.env

# build
RUN cmake -S /src -B /src/build \
    -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_TOOLCHAIN_FILE=/opt/vcpkg/scripts/buildsystems/vcpkg.cmake \
 && cmake --build /src/build

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/build/cpp-svc /app/cpp-svc
COPY --from=build /src/.env /app/.env

EXPOSE 8083

CMD ["/app/cpp-svc"]
