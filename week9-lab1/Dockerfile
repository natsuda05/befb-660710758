FROM golang:1.24.5 AS builder

#Copy Soure code ทุกอย่าง
WORKDIR /app
#COPY เพื่อหา Library ทีี่ต้องใช้
COPY go.mod go.sum ./
RUN go mod download 

#Copy Soure code ทุกอย่าง อีกรอบเพื่อมา Build
COPY . .  

#Build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

#Image ที่จะใช้งาน
FROM alpine:latest 

#ใช้ทดสอบยิง API เส้น  health
RUN apk --no-cache add ca-certificates curl

WORKDIR /root

COPY --from=builder /app/main .

ENTRYPOINT [ "./main"]