---
title: "Release Alerts"
weight: 60
aliases:
    /docs/reference/snsupdates/
description: >
  SNS Alerts for EKS Anywhere releases
---

EKS Anywhere uses Amazon Simple Notification Service (SNS) to notify availability of a new release.
It is recommended that your clusters are kept up to date with the latest EKS Anywhere release.
Please follow the instructions below to subscribe to SNS notification.

* Sign in to your AWS Account
* Select us-east-1 region
* Go to the SNS Console
* In the left navigation pane, choose "Subscriptions"
* On the *Subscriptions* page, choose "Create subscription"
* On the *Create subscription* page, in the *Details* section enter the following information
  * Topic ARN
    ```
    arn:aws:sns:us-east-1:153288728732:eks-anywhere-updates
    ```
  * Protocol - *Email*
  * Endpoint - *Your preferred email address*
* Choose *Create Subscription*
* In few minutes, you will receive an email asking you to confirm the subscription
* Click the confirmation link in the email
