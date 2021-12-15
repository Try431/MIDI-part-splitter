# syntax=docker/dockerfile:1

FROM python:3.8-slim-buster as build-image

ARG AWS_ACCESS_KEY_ID
ENV AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID

ARG AWS_SECRET_ACCESS_KEY
ENV AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY

# AWS Lambdas require a specific runtime environment, otherwise the lambda will error out no matter what
RUN pip install \
    --target /asset \
    awslambdaric

FROM ubuntu:20.04

WORKDIR /asset

COPY --from=build-image /asset /asset
COPY --from=build-image /usr/local/bin /usr/local/bin

# workaround to get past tzdata configuration input (https://archive.md/k3HbH)
ENV TZ=America/Chicago
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update
RUN apt-get install -y python3-pip fluidsynth ffmpeg

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY . .

ENTRYPOINT [ "/usr/local/bin/python", "-m", "awslambdaric" ]
CMD [ "midi_to_mp3_lambda.handler" ]