import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Divider from "@material-ui/core/Divider";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import Tooltip from "@material-ui/core/Tooltip";
import List from "@material-ui/core/List";

import { statusCodeToId } from "/lib/util/checkStatus";
import CardHighlight from "/lib/component/base/CardHighlight";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/lib/component/base/Dictionary";
import ListItem, { ListItemTitle } from "/lib/component/base/DetailedListItem";
import Maybe from "/lib/component/util/Maybe";
import { RelativeToCurrentDate } from "/lib/component/base/RelativeDate";
import StatusIcon from "/lib/component/base/CheckStatusIcon";
import SilencedIcon from "/lib/component/icon/Silence";

import LabelsAnnotationsCell from "/app/component/partial/LabelsAnnotationsCell";

const Strong = withStyles(() => ({
  root: {
    color: "inherit",
    fontSize: "inherit",
    fontWeight: 500,
  },
}))(Typography);

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

        metadata {
          ...LabelsAnnotationsCell_objectmeta
        }
      }
      ${LabelsAnnotationsCell.fragments.objectmeta}
    `,
  };

  render() {
    const {
      entity,
      entity: { system },
    } = this.props;
    const statusCode = entity.status;
    const status = statusCodeToId(statusCode);

    // Only display network interfaces that have a MAC address at
    // this time. This avoids displaying the loopback and tunnel
    // interfaces.
    const networkInterfaces = system.network.interfaces.filter(
      intr => intr.mac && intr.addresses.length > 0,
    );

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
                    {`${status} `}({statusCode})
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
              </Dictionary>
            </Grid>
          </Grid>
        </CardContent>
        <Divider />
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
        <Maybe value={networkInterfaces}>
          <Divider />
          <CardContent>
            <Grid container spacing={0}>
              {networkInterfaces.map(intr => (
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
              ))}
            </Grid>
          </CardContent>
        </Maybe>
        <Divider />
        <LabelsAnnotationsCell entity={entity} />
      </Card>
    );
  }
}

export default EntityDetailsInformation;
