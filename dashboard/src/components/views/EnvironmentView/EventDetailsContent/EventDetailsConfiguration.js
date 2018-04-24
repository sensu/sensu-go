import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card, { CardContent } from "material-ui/Card";
import Typography from "material-ui/Typography";
import Divider from "material-ui/Divider";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import RelativeDate from "/components/RelativeDate";
import Monospaced from "/components/Monospaced";
import Maybe from "/components/Maybe";

class EventDetailsConfiguration extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EventDetailsConfiguration_entity on Entity {
        name
        class
        system {
          platform
        }
        lastSeen
        subscriptions
      }
    `,
    check: gql`
      fragment EventDetailsConfiguration_check on Check {
        name
        command
        interval
        subscriptions
        timeout
        ttl
      }
    `,
  };

  render() {
    const { entity, check } = this.props;
    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Configuration
          </Typography>
          <Dictionary>
            <DictionaryEntry>
              <DictionaryKey>Entity</DictionaryKey>
              <DictionaryValue>{entity.name}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Class</DictionaryKey>
              <DictionaryValue>{entity.class}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Platform</DictionaryKey>
              <DictionaryValue>{entity.system.platform}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Last Seen</DictionaryKey>
              <DictionaryValue>
                <Maybe value={entity.lastSeen} fallback="unknown">
                  {val => <RelativeDate dateTime={val} />}
                </Maybe>
              </DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Subscriptions</DictionaryKey>
              <DictionaryValue>
                {entity.subscriptions.join(", ")}
              </DictionaryValue>
            </DictionaryEntry>
          </Dictionary>
        </CardContent>
        <Divider />
        <CardContent>
          <Dictionary>
            <DictionaryEntry>
              <DictionaryKey>Check</DictionaryKey>
              <DictionaryValue>{check.name}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Interval</DictionaryKey>
              <DictionaryValue>{check.interval}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Timeout</DictionaryKey>
              <DictionaryValue>{check.timeout}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>TTL</DictionaryKey>
              <DictionaryValue>{check.ttl}</DictionaryValue>
            </DictionaryEntry>
          </Dictionary>
        </CardContent>
        <Divider />
        <Monospaced background>
          <CardContent>{`# Executed command\n$ ${check.command}`}</CardContent>
        </Monospaced>
      </Card>
    );
  }
}

export default EventDetailsConfiguration;
