This page is used to list deprecation notices across GoReleaser.

Deprecated options are only removed on major versions of GoReleaser. Deprecated versions on features deemed experimental (you see a warning when using them) might be removed in feature releases.

Nevertheless, it’s a good thing to keep your configuration up-to-date to prevent any issues.

You can check your use of deprecated configurations by running:

```sh
goreleaser check
```

## Active deprecation notices

### dockers\_v2.retry

> since v2.15.3

All retry configuration was moved to the root of the configuration file:

### docker\_manifests.retry

> since v2.15.3

All retry configuration was moved to the root of the configuration file:

### dockers.retry

> since v2.15.3

All retry configuration was moved to the root of the configuration file:

### furies

> since v2.14

`furies` was renamed to `gemfury` for clarity.

### mcp.github

> since v2.13.1

The MCP configuration was initially nested under `github`, but the registry is actually run by the MCP committee, not GitHub specifically. The configuration should now be at the top level of `mcp`.

### homebrew\_casks.binary

> since v2.12.6

It should now be in plural form.

### dockers

We’re re-implementing the docker feature from the ground up.

The configuration now is way more concise, and the implementation is simpler as well.

Before, you would have to setup `dockers` and `docker_manifests`, now, only `dockers` (provisionally being called `dockers_v2`) is needed to achieve the same things.

Then, instead of building the images, pushing them, and then building the manifests and pushing them, we will now run a single `docker buildx build` with the given platforms, which will build and publish the manifest and SBOM.

As you can see, it’s a lot simpler. The resulting images are the same, a combination of all the non-empty images with all the non-empty tags.

This will also require a small change in your `Dockerfile` when copying from the context:

GoReleaser will automatically setup the context in such a way that all the artifacts for the given target platform will be located within `$TARGETPLATFORM/`.

When running a `--snapshot` build, GoReleaser will append the platform to the tags, and will create Docker images instead of manifests, as manifests can’t be created without pushing.

Feel free to suggest improvements [here](https://github.com/goreleaser/goreleaser/discussions/6005).

Regarding signing, you may also remove the `artifacts` option from you `docker_signs`:

Since in the future we’ll only have the docker image type, the `artifacts` property will eventually be deprecated and removed.

### homebrew\_casks.conflicts.formula

> since v2.12

It was a no-op before, and is now [removed from Homebrew](https://github.com/Homebrew/brew/pull/20499).

### homebrew\_casks.manpage

> since v2.11

You may now define multiple man pages, which was not possible in v2.10.

### brews

> since v2.10 (soft) since v2.16

Historically, GoReleaser would generate *hackyish* formulas that would install the pre-compiled binaries. This was the only way to do it for Linuxbrew at the time, but this is no longer true, and *Casks* should be used instead.

That said, we now have a `homebrew_casks` section!

For simple cases, simply replacing one with the other will be good enough. More complex settings might require further change. Check the [new documentation](https://goreleaser.com/customization/publish/homebrew_casks/) for more details.

Once you do the first release this way, you might also want to delete the old *Formulas* from your *Tap*. You may also want to make the *Cask* conflict with the previous *Formula*.

Warning

Don’t forget to remove the `directory: Formula` from your configuration. Casks **need** to be in the `Casks` directory - which is the default.

The preferred way to migrate is to create a `tap_migrations.json` file in the root of your tap:

```json
{
  "foo": "foo"
}
```

And then delete the Formula:

```sh
rm Formula/foo.rb
```

With this, when the user tries to upgrade, it should automatically update to the Cask instead.

### archives.builds

> since v2.8

The `builds` field has been replaced with the `ids`, which is the nomenclature used everywhere else.

### snaps.builds

> since v2.8

The `builds` field has been replaced with the `ids`, which is the nomenclature used everywhere else.

### nfpms.builds

> since v2.8

The `builds` field has been replaced with the `ids`, which is the nomenclature used everywhere else.

### archives.format

> since v2.6

Format was renamed to `formats`, and now accepts a list of formats.

Note

It will still accept a single string, e.g.: `formats: zip`. In most cases you can simply rename the property to formats.

### archives.format\_overrides.format

> since v2.6

Format was renamed to `formats`, and now accepts a list of formats.

Note

It will still accept a single string, e.g.: `formats: zip`. In most cases you can simply rename the property to formats.

Note

It will still accept a single string, e.g.: `formats: zip`. In most cases you can simply rename the property to formats.

### kos.repository

> since v2.5

Use `repositories` instead. It allows to create multiple images with Ko, without having to rebuild each of them.

### builds.gobinary

> since v2.5

The property was renamed to `tool`, as to better accommodate multiple languages.

### kos.sbom

> since v2.2

Ko removed support for `cyclonedx` and `go.version-m` SBOMs from upstream. You can now either use `spdx` or `none`. From now on, these two options will be replaced by `none`. We recommend you change it to `spdx`.

### nightly.name\_template

> since v2.2

Property renamed so its easier to reason about.

### snapshot.name\_template

> since v2.2

Property renamed so its easier to reason about.

## Removed in v2

### archives.strip\_parent\_binary\_folder

> since 2024-03-29 (v1.25), removed 2024-05-26 (v2.0)

Property was renamed to be consistent across all configurations.

### blobs.folder

> since 2024-03-29 (v1.25), removed 2024-05-26 (v2.0)

Property was renamed to be consistent across all configurations.

### brews.folder

> since 2024-03-29 (v1.25), removed 2024-05-26 (v2.0)

Property was renamed to be consistent across all configurations.

### scoops.folder

> since 2024-03-29 (v1.25), removed 2024-05-26 (v2.0)

Property was renamed to be consistent across all configurations.

### furies.skip

> since 2024-03-03 (v1.25), removed 2024-05-26 (v2.0)

Changed to `disable` to conform with all other pipes.

### changelog.skip

> since 2024-01-14 (v1.24), removed 2024-05-26 (v2.0)

Changed to `disable` to conform with all other pipes.

### blobs.kmskey

> since 2024-01-07 (v1.24), removed 2024-05-26 (v2.0)

Changed to `kms_key` to conform with all other options.

### blobs.disableSSL

> since 2024-01-07 (v1.24), removed 2024-05-26 (v2.0)

Changed to `disable_ssl` to conform with all other options.

### \--skip

> since 2023-09-14 (v1.21), removed 2024-05-26 (v2.0)

The following `goreleaser release` flags were deprecated:

- `--skip-announce`
- `--skip-before`
- `--skip-docker`
- `--skip-ko`
- `--skip-publish`
- `--skip-sbom`
- `--skip-sign`
- `--skip-validate`

By the same token, the following `goreleaser build` flags were deprecated:

- `--skip-before`
- `--skip-post-hooks`
- `--skip-validate`

All these flags are now under a single `--skip` flag, that accepts multiple values.

You can check `goreleaser build --help` and `goreleaser release --help` to see the valid options, and shell autocompletion should work properly as well.

### scoops.bucket

> since 2023-06-13 (v1.19.0), removed 2024-05-26 (v2.0)

Replace `bucket` with `repository`.

### krews.index

> since 2023-06-13 (v1.19.0), removed 2024-05-26 (v2.0)

Replace `index` with `repository`.

### brews.tap

> since 2023-06-13 (v1.19.0), removed 2024-05-26 (v2.0)

Replace `tap` with `repository`.

### archives.rlcp

> since 2023-06-06 (v1.19.0), removed 2024-05-26 (v2.0)

This option is now default and can’t be changed. You can remove it from your configuration files.

### source.rlcp

> since 2023-06-06 (v1.19.0), removed 2024-05-26 (v2.0)

This option is now default and can’t be changed. You can remove it from your configuration files.

### brews.plist

> since 2023-06-06 (v1.19.0), removed 2024-05-26 (v2.0)

`plist` is deprecated by Homebrew, and now on GoReleaser too. Use `service` instead.

### –debug

> since 2023-05-16 (v1.19.0), removed 2024-05-26 (v2.0)

`--debug` has been deprecated in favor of `--verbose`.

### scoop

> since 2023-04-30 (v1.18.0), removed 2024-05-26 (v2.0)

GoReleaser now allows many `scoop` configurations, so it should be pluralized [accordingly](https://goreleaser.com/customization/publish/scoop/).

### build

> since 2023-02-09 (v1.16.0), removed 2024-05-26 (v2.0)

This option was still being supported, even though undocumented, for a couple of years now. It’s finally time to sunset it.

Simply use the pluralized form, `builds`, according to the [documentation](https://goreleaser.com/customization/builds/).

### –rm-dist

> since 2023-01-17 (v1.15.0), removed 2024-05-26 (v2.0)

`--rm-dist` has been deprecated in favor of `--clean`.

### nfpms.maintainer

> since 2022-05-07 (v1.9.0), removed 2024-05-26 (v2.0)

nFPM will soon make mandatory setting the maintainer field.