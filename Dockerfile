FROM golang:1.21 as base

RUN go install github.com/cosmtrek/air@latest

# TODO freeze version?
FROM python as  nsfwmodel

RUN curl -LO https://github.com/GantMan/nsfw_model/releases/download/1.2.0/mobilenet_v2_140_224.1.zip && unzip mobilenet_v2_140_224.1.zip





