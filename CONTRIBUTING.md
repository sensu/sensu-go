## Contributing

We, the maintainers, love pull requests from everyone, but often find
we must say "no" despite how reasonable the proposal may seem.

For this reason, we ask that you open an issue to discuss proposed
changes prior to submitting a pull request for the implementation.
This helps us to provide direction as to implementation details, which
branch to base your changes on, and so on.

1. Open an issue to describe your proposed improvement or feature
2. Fork https://github.com/sensu/sensu-go and clone your fork to your workstation
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Commit your changes with a [DCO Signed-off-by statement](#dco) (`git commit --signoff`)
5. Push your feature branch (`git push origin my-new-feature`)
6. Create a Pull Request as appropriate based on the issue discussion

### DCO

To make a good faith effort to ensure the criteria of the MIT License
are met, Sensu Inc. requires the Developer Certificate of Origin (DCO)
process to be followed.

The DCO is an attestation attached to every contribution made by every
developer. In the commit message of the contribution, the developer
simply adds a Signed-off-by statement and thereby agrees to the DCO,
which you can find below or at http://developercertificate.org/.

```
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

The following is an example DCO Signed-off-by statement.

```
 Author: Sean Porter <sean@sensu.io>

 Committer: Greg Poirier <greg@sensu.io>

   Let's name it WizardFormat.

   Calling it the Sensu Metric Format was a mistake.

   Signed-off-by: Sean Porter <sean@sensu.io>
   Signed-off-by: Grep Poirier <greg@sensu.io>
```

Git makes this easy with `git commit --signoff`!

The DCO text can either be manually added to your commit body, or you
can add either `-s` or `--signoff` to your usual git commit commands.
If you forget to add the sign-off you can also amend a previous commit
with the sign-off by running `git commit --amend -s`. If you've pushed
your changes to Github already you'll need to force push your branch
after this with `git push -f`. -- Thanks Chef!
