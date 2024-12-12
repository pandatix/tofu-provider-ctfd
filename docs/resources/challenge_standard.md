---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ctfd_challenge_standard Resource - terraform-provider-ctfd"
subcategory: ""
description: |-
  CTFd is built around the Challenge resource, which contains all the attributes to define a part of the Capture The Flag event.
  It is the first historic implementation of its kind, with basic functionalities.
---

# ctfd_challenge_standard (Resource)

CTFd is built around the Challenge resource, which contains all the attributes to define a part of the Capture The Flag event.

It is the first historic implementation of its kind, with basic functionalities.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `category` (String) Category of the challenge that CTFd groups by on the web UI.
- `description` (String) Description of the challenge, consider using multiline descriptions for better style.
- `name` (String) Name of the challenge, displayed as it.
- `value` (Number) The value (points) of the challenge once solved.

### Optional

- `connection_info` (String) Connection Information to connect to the challenge instance, useful for pwn, web and infrastructure pentests.
- `max_attempts` (Number) Maximum amount of attempts before being unable to flag the challenge.
- `next` (Number) Suggestion for the end-user as next challenge to work on.
- `requirements` (Attributes) List of required challenges that needs to get flagged before this one being accessible. Useful for skill-trees-like strategy CTF. (see [below for nested schema](#nestedatt--requirements))
- `state` (String) State of the challenge, either hidden or visible.
- `tags` (List of String) List of challenge tags that will be displayed to the end-user. You could use them to give some quick insights of what a challenge involves.
- `topics` (List of String) List of challenge topics that are displayed to the administrators for maintenance and planification.

### Read-Only

- `id` (String) Identifier of the challenge.

<a id="nestedatt--requirements"></a>
### Nested Schema for `requirements`

Optional:

- `behavior` (String) Behavior if not unlocked, either hidden or anonymized.
- `prerequisites` (List of String) List of the challenges ID.