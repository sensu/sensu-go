import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import CronDescriptor from "/components/partials/CronDescriptor";
import Divider from "@material-ui/core/Divider";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import CodeBlock from "/components/CodeBlock";
import CodeHighlight from "/components/CodeHighlight/CodeHighlight";
import Maybe from "/components/Maybe";
import InlineLink from "/components/InlineLink";
import Typography from "@material-ui/core/Typography";
import Tooltip from "@material-ui/core/Tooltip";
import SilencedIcon from "/icons/Silence";
import Grid from "@material-ui/core/Grid";
import StatusIcon from "/components/CheckStatusIcon";
import List from "@material-ui/core/List";
import ListItem, { ListItemTitle } from "/components/DetailedListItem";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import ExpansionPanel from "@material-ui/core/ExpansionPanel";
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import { statusCodeToId } from "/utils/checkStatus";
import { withStyles } from "@material-ui/core/styles";
import { RelativeToCurrentDate } from "/components/RelativeDate";

const Strong = withStyles(() => ({
  root: {
    color: "inherit",
    fontSize: "inherit",
    fontWeight: 500,
  },
}))(Typography);

class EventDetailsSummary extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EventDetailsSummary_entity on Entity {
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
      entity: entityProp,
      entity: { system },
    } = this.props;
    const { namespace, ...entity } = entityProp;
    const statusCode = entity.status;
    const status = statusCodeToId(statusCode);

    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Entity Summary
            {entity.silences.length > 0 && (
              <Tooltip title="Silenced">
                <SilencedIcon style={{ float: "right" }} />
              </Tooltip>
            )}
          </Typography>
          <Grid container spacing={0}>
            <Grid item xs={12} sm={6}>
              <Dictionary>
                <DictionaryEntry>
                  <DictionaryKey>Entity</DictionaryKey>
                  <DictionaryValue>{entity.name}</DictionaryValue>
                </DictionaryEntry>
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
                  <DictionaryKey>User</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={entity.user} fallback="—" />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Deregister</DictionaryKey>
                  <DictionaryValue>
                    {entity.deregister ? "yes" : "no"}
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Deregistration Handler</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={system.deregistration} fallback="—">
                      {config => config.handler}
                    </Maybe>
                  </DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
          </Grid>
        </CardContent>
        <ExpansionPanel>
          <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
            <Typography variant="button">More</Typography>
          </ExpansionPanelSummary>
          <ExpansionPanelDetails>
            <CardContent>
              <Grid container spacing={0}>
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
                  </Dictionary>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Dictionary>
                    <DictionaryEntry>
                      <DictionaryKey>Redacted Keys</DictionaryKey>
                      <DictionaryValue>
                        {entity.redact.length > 0 ? (
                          <List disablePadding>
                            {entity.redact.map(key => (
                              <ListItem key={key}>
                                <ListItemTitle>{key}</ListItemTitle>
                              </ListItem>
                            ))}
                          </List>
                        ) : (
                          "—"
                        )}
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
                            [
                              `${system.platform} ${system.platformVersion}`,
                              system.platformFamily,
                            ]
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
                            <DictionaryKey>Adapter</DictionaryKey>
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
          </ExpansionPanelDetails>
        </ExpansionPanel>
      </Card>
    );
  }
}

export default EventDetailsSummary;
