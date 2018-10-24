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
import { RelativeToCurrentDate } from "/components/RelativeDate";
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
        ns: namespace {
          org: organization
          env: environment
        }
        system {
          platform
        }

        name
        class
        lastSeen
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
    const { ns, ...entity } = entityProp;

    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Check Summary
          </Typography>
          <Dictionary>
            <DictionaryEntry>
              <DictionaryKey>Check</DictionaryKey>
              <DictionaryValue>
                <InlineLink to={`/${ns.org}/${ns.env}/checks/${check.name}`}>
                  {check.name}
                </InlineLink>
              </DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Schedule</DictionaryKey>
              <DictionaryValue>
                <Maybe value={check.cron} fallback={`${check.interval}s`}>
                  {cron => <CronDescriptor capitalize expression={cron} />}
                </Maybe>
              </DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>Timeout</DictionaryKey>
              <DictionaryValue>{check.timeout}s</DictionaryValue>
            </DictionaryEntry>
            <DictionaryEntry>
              <DictionaryKey>TTL</DictionaryKey>
              <DictionaryValue>{check.ttl}s</DictionaryValue>
            </DictionaryEntry>
          </Dictionary>
        </CardContent>
        <Divider />
        <CardContent>
          <Dictionary>
            <DictionaryEntry>
              <DictionaryKey>Entity</DictionaryKey>
              <DictionaryValue>
                <InlineLink to={`/${ns.org}/${ns.env}/entities/${entity.name}`}>
                  {entity.name}
                </InlineLink>
              </DictionaryValue>
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
                  {val => <RelativeToCurrentDate dateTime={val} />}
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
        {check.command && (
          <React.Fragment>
            <Divider />
            <CodeBlock>
              <CardContent>
                <CodeHighlight
                  language="bash"
                  code={`# Executed command\n$ ${check.command}`}
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

export default EventDetailsSummary;
