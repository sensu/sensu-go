import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "material-ui/styles";
import Card, { CardContent } from "material-ui/Card";
import Divider from "material-ui/Divider";
import Grid from "material-ui/Grid";
import Typography from "material-ui/Typography";
import { statusCodeToId } from "/utils/checkStatus";
import CardHighlight from "/components/CardHighlight";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import Maybe from "/components/Maybe";
import Monospaced from "/components/Monospaced";
import RelativeDate from "/components/RelativeDate";
import StatusIcon from "/components/CheckStatusIcon";

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
        class
        subscriptions
        keepaliveTimeout
        lastSeen
        status
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
          <Typography variant="headline">{entity.name}</Typography>
          <Typography variant="caption" paragraph>
            {entity.subscriptions.join(", ")}
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
                <DictionaryEntry>
                  <DictionaryKey>Last Seen</DictionaryKey>
                  <DictionaryValue>
                    <RelativeDate dateTime={entity.lastSeen} />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Class</DictionaryKey>
                  <DictionaryValue>{entity.class}</DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>User</DictionaryKey>
                  <DictionaryValue>{entity.user}</DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Dictionary>
                <DictionaryEntry>
                  <DictionaryKey>Hostname</DictionaryKey>
                  <DictionaryValue>{system.hostname}</DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>OS</DictionaryKey>
                  <DictionaryValue>{system.os}</DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Platform</DictionaryKey>
                  <DictionaryValue>
                    {system.platform} / {system.platformFamily}
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Architecture</DictionaryKey>
                  <DictionaryValue>{system.arch}</DictionaryValue>
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
            <Monospaced background>
              <CardContent>
                {JSON.stringify(entity.extendedAttributes, null, "\t")}
              </CardContent>
            </Monospaced>
          </React.Fragment>
        )}
      </Card>
    );
  }
}

export default EntityDetailsInformation;
