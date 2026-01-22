FROM maven:3.9-eclipse-temurin-17 AS build
WORKDIR /src
COPY . .
RUN mvn -q -DskipTests package

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=build /src/target/java-service-1.0.0.jar /app/app.jar
EXPOSE 8080 9090
ENTRYPOINT ["java","-jar","/app/app.jar"]
