import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card, { CardContent } from "material-ui/Card";
import Divider from "material-ui/Divider";
import Typography from "material-ui/Typography";
import Grid from "material-ui/Grid";
import { statusCodeToId } from "/utils/checkStatus";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import CardHighlight from "/components/CardHighlight";
import RelativeDate from "/components/RelativeDate";
import {
  DateStringFormatter,
  DateTime,
  KitchenTime,
} from "/components/DateFormatter";
import Duration from "/components/Duration";
import StatusIcon from "/components/CheckStatusIcon";
import Monospaced from "/components/Monospaced";
import Maybe from "/components/Maybe";

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
                  <DictionaryKey>Last OK</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={check.lastOK} fallback="Never">
                      {val => <RelativeDate dateTime={val} capitalize />}
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
                  <DictionaryValue>{check.name}</DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Entity</DictionaryKey>
                  <DictionaryValue>{entity.name}</DictionaryValue>
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
            <Monospaced background>
              <CardContent>{check.output}</CardContent>
            </Monospaced>
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
