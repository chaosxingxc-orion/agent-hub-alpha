FROM eclipse-temurin:21-jdk AS build
WORKDIR /src
# 纯 JDK 部分：契约（剥离 jackson 注解由构建脚本完成）+ checker + http server
# 为保持镜像最小，运行时用 jre
COPY _pure/ /src/
RUN mkdir -p out && find . -name "*.java" | sort > srcs.txt \
    && javac -d out @srcs.txt

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /src/out /app/classes
# 快照目录由 compose 挂载（W01 冻结材料）；演示期先用内置样例快照
COPY snapshot/ /snapshot/
ENV CHECKER_PORT=18095 SNAPSHOT_DIR=/snapshot
EXPOSE 18095
HEALTHCHECK --interval=15s --timeout=5s --retries=5 \
  CMD wget -q -O- http://localhost:18095/health || exit 1
ENTRYPOINT ["java", "-cp", "/app/classes", "com.agenthub.w06.checker.server.CheckerHttpServer"]
