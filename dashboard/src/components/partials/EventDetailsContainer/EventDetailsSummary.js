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

class EventDetailsSummary extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    entity: gql`
      fragment EventDetailsSummary_entity on Entity {
        namespace
        system {
          platform
        }

        name
        entityClass
        subscriptions
      }
    `,
    check: gql`
      fragment EventDetailsSummary_check on Check {
        name
        command
        interval
        cron
        subscriptions
        timeout
        ttl
      }
    `,
  };

  render() {
    const { entity: entityProp, check } = this.props;
    const { namespace, ...entity } = entityProp;

    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Entity Summary
          </Typography>
          <Dictionary>
            <DictionaryEntry>
              <DictionaryKey>Entity</DictionaryKey>
              <DictionaryValue>
                <InlineLink to={`/${namespace}/entities/${entity.name}`}>
                  {entity.name}
                </InlineLink>
              </DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Class</DictionaryKey>
              <DictionaryValue>{entity.entityClass}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Platform</DictionaryKey>
              <DictionaryValue>{entity.system.platform}</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Subscriptions</DictionaryKey>
              <DictionaryValue>
                {entity.subscriptions.join(", ")}
              </DictionaryValue>
            </DictionaryEntry>
          </Dictionary>
        </CardContent>
      </Card>
    );
  }
}

export default EventDetailsSummary;
