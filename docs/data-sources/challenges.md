---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ctfd_challenges Data Source - terraform-provider-ctfd"
subcategory: ""
description: |-
  
---

# ctfd_challenges (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `challenges` (Attributes List) (see [below for nested schema](#nestedatt--challenges))
- `id` (String) The ID of this resource.

<a id="nestedatt--challenges"></a>
### Nested Schema for `challenges`

Read-Only:

- `category` (String) Category of the challenge that CTFd groups by on the web UI.
- `connection_info` (String) Connection Information to connect to the challenge instance, useful for pwn or web pentest.
- `decay` (Number)
- `description` (String) Description of the challenge, consider using multiline descriptions for better style.
- `files` (Attributes List) List of files given to players to flag the challenge. (see [below for nested schema](#nestedatt--challenges--files))
- `flags` (Attributes List) List of challenge flags that solves it. (see [below for nested schema](#nestedatt--challenges--flags))
- `function` (String) Decay function to define how the challenge value evolve through solves, either linear or logarithmic.
- `hints` (Attributes List) List of hints about the challenge displayed to the end-user. (see [below for nested schema](#nestedatt--challenges--hints))
- `id` (String) Identifier of the challenge.
- `initial` (Number)
- `max_attempts` (Number) Maximum amount of attempts before being unable to flag the challenge.
- `minimum` (Number)
- `name` (String) Name of the challenge, displayed as it.
- `requirements` (Attributes) List of required challenges that needs to get flagged before this one being accessible. Useful for skill-trees-like strategy CTF. (see [below for nested schema](#nestedatt--challenges--requirements))
- `state` (String) State of the challenge, either hidden or visible.
- `tags` (List of String) List of challenge tags that will be displayed to the end-user. You could use them to give some quick insights of what a challenge involves.
- `topics` (List of String) List of challenge topics that are displayed to the administrators for maintenance and planification.
- `type` (String) Type of the challenge defining its layout, either standard or dynamic.
- `value` (Number)

<a id="nestedatt--challenges--files"></a>
### Nested Schema for `challenges.files`

Read-Only:

- `content` (String)
- `contentb64` (String)
- `id` (String)
- `location` (String)
- `name` (String)


<a id="nestedatt--challenges--flags"></a>
### Nested Schema for `challenges.flags`

Read-Only:

- `content` (String)
- `data` (String)
- `id` (String)
- `type` (String)


<a id="nestedatt--challenges--hints"></a>
### Nested Schema for `challenges.hints`

Read-Only:

- `content` (String)
- `cost` (Number)
- `id` (String)
- `requirements` (List of String)


<a id="nestedatt--challenges--requirements"></a>
### Nested Schema for `challenges.requirements`

Read-Only:

- `behavior` (String) Behavior if not unlocked, either hidden or anonymized.
- `prerequisites` (List of String) List of the challenges ID.