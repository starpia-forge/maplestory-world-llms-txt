# MapleStory World LLM Docs

## Overview

A project that processes `MapleStory World` development guides and API Reference documentation into markdown documents
that are easy for LLMs to understand.

Includes the implementation of a web crawler for this purpose. For those who want to check the documents, please refer
to the [Documents](#documents) section.

## Documents

- [English Documents](/docs/en)
- [Korean Documents](/docs/kr)

## AI Assistants

The documents in this repository will be available for use in the dedicated AI agents below.

| Assistant            | Platform      | Link       |
|:---------------------|:--------------|:-----------|
| **MapleStory World** | ChatGPT (GPT) | [Link](링크) |
| **MapleStory World** | Gemini (Gem)  | [Link](링크) |

## License

- **Documents**: The documents included in this project are processed content scraped from the [References](#references)
  sites. Therefore, the copyright of the documents belongs to `Nexon`, and distribution may be discontinued at any time
  upon their request.
  The project used for document processing is `mdream`, which is licensed under the MIT License.
- **Source Code**: The crawler created for site scraping is licensed under the [MIT License](/LICENSE).

## References

- [MapleStory World (English)](https://maplestoryworlds-creators.nexon.com/en)
- [MapleStory World (Korean)](https://maplestoryworlds-creators.nexon.com/ko)
- [mdream](https://github.com/harlan-zw/mdream)