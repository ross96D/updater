#!/bin/python3.10
import subprocess
import argparse
from typing import List, cast
from functools import total_ordering


@total_ordering
class Version:
    def __init__(self, major: int, minor: int, patch: int) -> None:
        self.major = major
        self.minor = minor
        self.patch = patch

    def __str__(self) -> str:
        return f"v{self.major}.{self.minor}.{self.patch}"

    def __eq__(self, other: object) -> bool:  # type: ignore
        if other is Version:
            other = cast(Version, other)
            return (
                self.major == other.major
                and self.minor == other.minor
                and self.patch == other.patch
            )
        return False

    def __lt__(self, other: "Version") -> bool:
        if self.major < other.major:
            return True
        elif self.major > other.major:
            return False

        if self.minor < other.minor:
            return True
        elif self.minor > other.minor:
            return False

        if self.patch < other.patch:
            return True
        elif self.patch > other.patch:
            return False

        return False

    @staticmethod
    def fromString(tag: str):
        assert tag[0] == "v"
        tag = tag.strip("v")
        splitted = tag.split(".")

        if len(splitted) != 3:
            raise Exception(
                f"version is not setted correctly, should be something like this v1.3.12 and is {tag}"
            )

        return Version(
            int(splitted[0]),
            int(splitted[1]),
            int(splitted[2]),
        )


def up_version(v: Version):
    with open("./share/version", "w") as f:
        f.write(v.__str__())

    p = subprocess.run(["git", "add", "share/version"], check=True)
    assert p.returncode == 0
    p = subprocess.run(["git", "commit", "-m", f"bump version to {v}"], check=True)
    assert p.returncode == 0
    p = subprocess.run(["git", "tag", f"{v}"], check=True)
    assert p.returncode == 0
    p = subprocess.run(["git", "push"], check=True)
    assert p.returncode == 0
    p = subprocess.run(["git", "push", "origin", f"{v}"], check=True)
    assert p.returncode == 0


def majorFunc(v: Version):
    v.major += 1
    v.minor = 0
    v.patch = 0
    print("Major", v.__str__(), "to", v)
    up_version(v)


def minorFunc(v: Version):
    v.minor += 1
    v.patch = 0
    print("Minor", v.__str__(), "to", v)
    up_version(v)


def patchFunc(v: Version):
    v.patch += 1
    print("Patch", v.__str__(), "to", v)
    up_version(v)


def parse_arguments(lastTag: Version):
    parser = argparse.ArgumentParser(
        prog="./upver.py",
        description="bump version",
    )

    subparsers = parser.add_subparsers(required=True)

    major = subparsers.add_parser("major", description="up major version")
    major.set_defaults(func=majorFunc)

    minor = subparsers.add_parser("minor", description="up minor version")
    minor.set_defaults(func=minorFunc)

    patch = subparsers.add_parser("patch", description="up patch version")
    patch.set_defaults(func=patchFunc)

    args = parser.parse_args()

    args.func(lastTag)


def getTags() -> List[Version]:
    p = subprocess.run(
        ["git", "tag", "--list"],
        stdout=subprocess.PIPE,
        check=True,
        text=True,
    )

    lines = p.stdout.split("\n")
    return parseTages(lines)


def parseTages(lines: list[str]) -> List[Version]:
    result: list[Version] = []

    for s in lines:
        if s == "":
            continue
        result.append(Version.fromString(s))

    result.sort()
    return result


if __name__ == "__main__":
    tags = getTags()

    if len(tags) == 0:
        lastVersion = Version.fromString("v0.0.0")
    else:
        lastVersion = tags[-1]

    parse_arguments(lastVersion)
