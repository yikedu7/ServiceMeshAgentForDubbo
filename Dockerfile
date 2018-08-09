# Builder container
FROM registry.cn-shenzhen.aliyuncs.com/runningman/go-etcd AS builder1
COPY . $GOPATH/src/code.aliyun.com/runningguys/agent
WORKDIR $GOPATH/src/code.aliyun.com/runningguys/agent
RUN go get -v && go build . && cp  agent test.sh start-agent.sh /tmp

FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/services AS builder2

# Runner container
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/debian-jdk8
#COPY --from=builder /root/workspace/agent/agent /usr/local/bin/
COPY --from=builder1 /tmp/start-agent.sh /usr/local/bin
COPY --from=builder1 /tmp/agent /usr/local/bin
COPY --from=builder1 /tmp/test.sh /usr/local/bin

COPY --from=builder2 /root/workspace/services/mesh-provider/target/mesh-provider-1.0-SNAPSHOT.jar /root/dists/mesh-provider.jar
COPY --from=builder2 /root/workspace/services/mesh-consumer/target/mesh-consumer-1.0-SNAPSHOT.jar /root/dists/mesh-consumer.jar

COPY --from=builder2 /usr/local/bin/docker-entrypoint.sh /usr/local/bin

RUN set -ex && mkdir -p /root/logs

EXPOSE 8087

ENTRYPOINT ["docker-entrypoint.sh"]