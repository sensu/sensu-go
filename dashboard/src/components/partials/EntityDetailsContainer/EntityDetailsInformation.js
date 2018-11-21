import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Divider from "@material-ui/core/Divider";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { statusCodeToId } from "/utils/checkStatus";
import CardHighlight from "/components/CardHighlight";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import List from "@material-ui/core/List";
import ListItem, { ListItemTitle } from "/components/DetailedListItem";
import Maybe from "/components/Maybe";
import CodeBlock from "/components/CodeBlock";
import CodeHighlight from "/components/CodeHighlight/CodeHighlight";
import { RelativeToCurrentDate } from "/components/RelativeDate";
import StatusIcon from "/components/CheckStatusIcon";
import SilencedIcon from "/icons/Silence";
import Tooltip from "@material-ui/core/Tooltip";

const Strong = withStyles({
  root: {
    color: "inherit",
    fontSize: "inherit",
    fontWeight: 500,
  },
})(Typography);

class EntityDetailsInformation extends React.PureComponent {
  static propTypes = {
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsInformation_entity on Entity {
        name
        entityClass
        subscriptions
        lastSeen
        status
        silences {
          name
        }
        user
        redact
        extendedAttributes
        deregister
        deregistration {
          handler
        }

        system {
          arch
          os
          hostname
          platform
          platformFamily
          platformVersion

          network {
            interfaces {
              name
              addresses
              mac
            }
          }
        }
      }
    `,
  };

  render() {
    const {
      entity,
      entity: { system },
    } = this.props;
    const statusCode = entity.status;
    const status = statusCodeToId(statusCode);

    return (
      <Card>
        <CardHighlight color={status} />
        <CardContent>
          <Typography variant="headline">
            {entity.name}
            {entity.silences.length > 0 && (
              <Tooltip title="Silenced">
                <SilencedIcon style={{ float: "right" }} />
              </Tooltip>
            )}
          </Typography>
          <Typography variant="caption" paragraph>
            Current state of the {entity.entityClass}.
          </Typography>
          <Grid container spacing={0}>
            <Grid item xs={12} sm={6}>
              <Dictionary>
                <DictionaryEntry>
                  <DictionaryKey>Status</DictionaryKey>
                  <DictionaryValue>
                    <StatusIcon inline small statusCode={statusCode} />{" "}
                    {`${status} `}
                    ({statusCode})
                  </DictionaryValue>
                </DictionaryEntry>
                {entity.silences.length > 0 && (
                  <DictionaryEntry>
                    <DictionaryKey>Silenced By</DictionaryKey>
                    <DictionaryValue>
                      {entity.silences.map(s => s.name).join(", ")}
                    </DictionaryValue>
                  </DictionaryEntry>
                )}
                <DictionaryEntry>
                  <DictionaryKey>Last Seen</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={entity.lastSeen} fallback="n/a">
                      {val => <RelativeToCurrentDate dateTime={val} />}
                    </Maybe>
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>User</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={entity.user} fallback="n/a" />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Subscriptions</DictionaryKey>
                  <DictionaryValue>
                    {entity.subscriptions.length > 0 ? (
                      <List disablePadding>
                        {entity.subscriptions.map(subscription => (
                          <ListItem key={subscription}>
                            <ListItemTitle>{subscription}</ListItemTitle>
                          </ListItem>
                        ))}
                      </List>
                    ) : (
                      "—"
                    )}
                  </DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Dictionary>
                <DictionaryEntry>
                  <DictionaryKey>Class</DictionaryKey>
                  <DictionaryValue>{entity.entityClass}</DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Hostname</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={system.hostname} fallback="n/a" />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>OS</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={system.os} fallback="n/a" />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Platform</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={system.platform} fallback="n/a">
                      {() =>
                        [system.platform, system.platformFamily]
                          .reduce(
                            (memo, val) => (val ? [...memo, val] : memo),
                            [],
                          )
                          .join(" / ")
                      }
                    </Maybe>
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Architecture</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={system.arch} fallback="n/a" />
                  </DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
          </Grid>
        </CardContent>
        <Divider />
        <CardContent>
          <Grid container spacing={0}>
            {system.network.interfaces.map(
              intr =>
                // Only display network interfaces that have a MAC address at
                // this time. This avoids displaying the loopback and tunnel
                // interfaces.
                intr.mac &&
                intr.addresses.length > 0 && (
                  <Grid item xs={12} sm={6} key={intr.name}>
                    <Dictionary>
                      <DictionaryEntry>
                        <DictionaryKey>&nbsp;</DictionaryKey>
                        <DictionaryValue>
                          <Strong>{intr.name}</Strong>
                        </DictionaryValue>
                      </DictionaryEntry>
                      <DictionaryEntry>
                        <DictionaryKey>MAC</DictionaryKey>
                        <DictionaryValue>
                          <Maybe value={intr.mac} fallback={"n/a"} />
                        </DictionaryValue>
                      </DictionaryEntry>
                      {intr.addresses.map((address, i) => (
                        <DictionaryEntry key={address}>
                          <DictionaryKey>
                            {i === 0 ? "IP Address" : <span>&nbsp;</span>}
                          </DictionaryKey>
                          <DictionaryValue>{address}</DictionaryValue>
                        </DictionaryEntry>
                      ))}
                    </Dictionary>
                  </Grid>
                ),
            )}
          </Grid>
        </CardContent>
        {Object.keys(entity.extendedAttributes).length > 0 && (
          <React.Fragment>
            <Divider />
            <CodeBlock>
              <CardContent>
                <CodeHighlight
                  language="json"
                  code={JSON.stringify(entity.extendedAttributes, null, "\t")}
                >
                  {code => <code dangerouslySetInnerHTML={{ __html: code }} />}
                </CodeHighlight>
              </CardContent>
            </CodeBlock>
          </React.Fragment>
        )}
      </Card>
    );
  }
}

export default EntityDetailsInformation;
