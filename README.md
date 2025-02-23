# deLLMiter

![deLLMiter](./logo.png)

🚧 Warning: deLLMiter is a work-in-progress prototype. Be prepared for rapid updates and changes in documentation, architecture, and functionality over the next few weeks as we continue to develop and refine this project. Contributions are welcomed! 🚧

**deLLMiter** is engineered to identify and detect LLM-specific delimiters, which can be undocumented and serve as a potential attack vector.

## Context

This project builds upon the general ideas presented in the following articles:
* [LLM Delimiters and Higher-Order Expressions](https://glthr.com/llm-delimiters-and-higher-order-expressions)
* [First-Order, Second-Order Expressions, and Delimiters in Languages](https://glthr.com/first-order-second-order-expressions-and-delimiters-in-languages)

## Operating Mode

At the current stage of development, deLLMiter generates messages containing first-order and higher-order expressions blended with a set of predefined delimiters (from `known_delimiters.txt`) commonly used across multiple models (we will complete this list soon). It then feeds these messages to the model, instructing it to respond *verbatim*.

In this initial phase, any discrepancies between input and output are stored in `./results`, and the system crudely detects potential delimiter usage by identifying their absence in the response. Future iterations will refine this process by mutating known delimiters and introducing new ones (notably through escape sequences).

# Demo

https://github.com/user-attachments/assets/3aa0a74f-5bf2-41db-aded-b6962199960b

# Usage

## Prerequisite

To use deLLMiter, a locally running LLM server is needed.

We recommend using [LM Studio](https://lmstudio.ai/) and setting it up [as a local LLM API server](https://lmstudio.ai/docs/api).

## Run

Run deLLMiter:

```bash
$ go run main.go -model {model_name}
```

Example:
```bash
$ go run main.go -model llama-3.2-3b-instruct
```

Any discrepancy between the message sent to the model and the response is logged in `./results/{model_name}_all.txt`

If delimiters are found, they are logged in  `./results/{model_name}_delimiters.txt`.
