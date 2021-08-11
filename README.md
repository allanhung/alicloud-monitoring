# Alicloud monitoring running on kubernetes

Version: 1.0.15

## Pre-Requirement
### RAM Permissions
```
{
    "Version": "1",
    "Statement": [
        {
            "Action": [
                "ecs:AddTag*"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "ecs:Describe*"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "vpc:Describe*"
            ],
            "Resource": "*",
            "Effect": "Allow"
        }
    ]
}
```
### Setup [Kube2ram](https://github.com/allanhung/kube2ram/tree/go-mod)

## Deploy
```
bumpversion patch
```

### Commands:
* updatek8stags: ECS tag update for kubernetes worker
* spotprice: spot instance price for kubernetes worker
