import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Divider from "@material-ui/core/Divider";
import Typography from "@material-ui/core/Typography";
import Grid from "@material-ui/core/Grid";
import { statusCodeToId } from "/utils/checkStatus";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import CardHighlight from "/components/CardHighlight";
import { RelativeToCurrentDate } from "/components/RelativeDate";
import {
  DateStringFormatter,
  DateTime,
  KitchenTime,
} from "/components/DateFormatter";
import Duration from "/components/Duration";
import StatusIcon from "/components/CheckStatusIcon";
import CodeBlock from "/components/CodeBlock";
import Maybe from "/components/Maybe";
import SilencedIcon from "/icons/Silence";
import Tooltip from "@material-ui/core/Tooltip";

import NamespaceLink from "/components/util/NamespaceLink";
import InlineLink from "/components/InlineLink";

class EventDetailsCheckResult extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object.isRequired,
    entity: PropTypes.object.isRequired,
  };

  static fragments = {
    check: gql`
      fragment EventDetailsCheckResult_check on Check {
        status
        lastOK
        occurrences
        occurrencesWatermark
        name
        executed
        issued
        duration
        output
        silenced
      }
    `,
    entity: gql`
      fragment EventDetailsCheckResult_entity on Entity {
        name
      }
    `,
  };

  render() {
    const { check, entity } = this.props;
    const statusCode = check.status;
    const status = statusCodeToId(check.status);
    const formatter = new Intl.NumberFormat("en-US");

    return (
      <Card>
        <CardHighlight color={status} />
        <CardContent>
          <Typography variant="headline" paragraph>
            Check Result
            {check.silenced.length > 0 && (
              <Tooltip title="Silenced">
                <SilencedIcon style={{ float: "right" }} />
              </Tooltip>
            )}
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
                {check.silenced.length > 0 && (
                  <DictionaryEntry>
                    <DictionaryKey>Silenced By</DictionaryKey>
                    <DictionaryValue>
                      {check.silenced.join(", ")}
                    </DictionaryValue>
                  </DictionaryEntry>
                )}
                <DictionaryEntry>
                  <DictionaryKey>Last OK</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={check.lastOK} fallback="Never">
                      {val => (
                        <RelativeToCurrentDate dateTime={val} capitalize />
                      )}
                    </Maybe>
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Occurrences</DictionaryKey>
                  <DictionaryValue>
                    {formatter.format(check.occurrences)}
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Max Occurrences</DictionaryKey>
                  <DictionaryValue>
                    {formatter.format(check.occurrencesWatermark)}
                  </DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Dictionary>
                <DictionaryEntry>
                  <DictionaryKey>Check</DictionaryKey>
                  <DictionaryValue>
                    <NamespaceLink
                      component={InlineLink}
                      to={`/checks/${check.name}`}
                    >
                      {check.name}
                    </NamespaceLink>
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Entity</DictionaryKey>
                  <DictionaryValue>
                    <NamespaceLink
                      component={InlineLink}
                      to={`/entities/${entity.name}`}
                    >
                      {entity.name}
                    </NamespaceLink>
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Issued at</DictionaryKey>
                  <DictionaryValue>
                    <DateStringFormatter
                      component={DateTime}
                      dateTime={check.issued}
                      short
                    />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Ran at</DictionaryKey>
                  <DictionaryValue>
                    <DateStringFormatter
                      component={KitchenTime}
                      dateTime={check.executed}
                    />
                    {" for "}
                    <Duration duration={check.duration * 1000} />
                  </DictionaryValue>
                </DictionaryEntry>
              </Dictionary>
            </Grid>
          </Grid>
        </CardContent>

        {check.output ? (
          <React.Fragment>
            <Divider />
            <CodeBlock>
              <CardContent>{check.output}</CardContent>
            </CodeBlock>
          </React.Fragment>
        ) : (
          <React.Fragment>
            <Divider />
            <CardContent>
              <Typography color="textSecondary" align="center">
                Check did not write to STDOUT.
              </Typography>
            </CardContent>
          </React.Fragment>
        )}
      </Card>
    );
  }
}

export default EventDetailsCheckResult;
