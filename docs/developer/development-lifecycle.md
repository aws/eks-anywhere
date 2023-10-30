# Development lifecycle

This documents aims to codify our development process for the most common scenarios from the moment you are assigned to a task (generally tied to a GitHub issue, whether it's a new feature, improvement or bugfix) to the moment when you can call the task *done*.

## Process

1. Read the GitHub issue and try to answer the following questions:
   1. What problem are we solving? This should be clear most of the time but if it's not, you might need to talk to the author of the issue and/or the stakeholders.
   1. Do we need to solve this problem and if so, why? If a feature advanced all the way to a GitHub issue, the answer most of the time will be yes. But don't fear to question the requirements. It's important to understand not only what we are building but why, or we might end up implementing the wrong thing. If you don't think this problem needs to be solved, it's already been solved or there a better way, this is the moment to voice it out.
   2. What are we building? Sometimes the issue will be prescriptive about the solution to the problem, sometimes it will just give you functional requirements and other times you will only have a description of the problem. Make sure you understand all of this and collect all the requirements before moving forward with the design. This will sometimes involve talking to other engineers, product or even users. If you disagree with any given requirement, challenge them.
1. If the task needs a design, work on it before starting the implementation.
   * Not all features need a design doc, but most will need a design. Study the problem space and existing code and come up with a solution before implementing it.
   * When possible, favor getting eyes on your ideas/design/doc as soon as possible. Early feedback avoids rework and will set you up for a speedy review once you write the code. Not everything needs a formal design review on a PR. Just ask folks for a chat to bounce ideas, especially folks with special experience or interest in the are you are working on. This is important when introducing new ideas or patterns.
   * If you are basing your design on one or more ideas you are not sure will work, feel free to play with a POC. The goal of designing the solution before implementing is to avoid rework and get early feedback, code is as valid as any other tool to come up with that design. That said, most features won't need a POC.
   * Timebox this, specially for big features. If it's taking too long (relative to the size of the task) you might be trying to go too deep. In that case, ask someone to pair on the design, it's usually easier to stay focused on what is needed when you can keep each other accountable.
   * Favor results over perfection, done is better than perfect. Look for two way doors and if a decision can be postponed without adding risk, leave it for later. At some point, you need to say "this is enough" and move forward with the implementation, even if that means making mistakes and having to solve them along the way.
   * Make sure you include testing in your design. When a feature can benefit from E2E tests, it can be useful to enumerate the scenarios that need to be tested.
1. If needed, split the task into manageable chunks. Feel free to create new GitHub issues. You can track them in the original one with a checklist or you can even create a GitHub project (specially useful for big features, even more so when they span across multiple releases). This helps to keep track of the progress and facilitates adding more people to the feature if needed.
   * When the task becomes a project, try to think into parallelizable work streams, specially when you know you'll be working with more people.
   * Make sure you account for documentation when needed.
   * Follow [this](issues.md) to categorize your issues.
1. Come up with an estimate of how long the task will take. You won't always need to commit to a date or even communicate it externally, but it helps keeping track of the progress and keeps us accountable.
1. Implement your design. Reading our [development](best-practice.md) and [PR best practices](pull-request.md) is a must.
   * Keep the team informed of your progress. If you get stuck, ask for help.
   * If you think the delivery on time of the task is in risk, bring it up to the team. In this kind of situation, you can generally compromise on time, quality, scope or resources, even though the latter tends to work very rarely. Make sure the team feels comfortable with the compromise (or combination of them) and plan to account for the consequences (communicating a delay to external stakeholders, addressing tech debt, communicating a scope reduction, etc.).

## When is a task *done*?

* The solution has been implemented, it solves the original issue and follows all requirements.
* All PRs have been merged.
* All new code is properly tested.
* If E2E tests were necessary, they are merged and running in CI (ideally without flakes).
* Documentation has been added (when necessary).
