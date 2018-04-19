import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card, { CardContent } from "material-ui/Card";
import Divider from "material-ui/Divider";
import Typography from "material-ui/Typography";
import Grid from "material-ui/Grid";

import { statusCodeToId, statusToColor } from "../../../../utils/checkStatus";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "../../../Dictionary";
import CardHighlight from "../../../CardHighlight";
import RelativeDate from "../../../RelativeDate";
import { DateTime, KitchenTime } from "../../../DateFormatter";
import Duration from "../../../Duration";
import StatusIcon from "../../../CheckStatusIconSmall";
import Monospaced from "../../../Monospaced";

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
    const statusText = statusCodeToId(check.status);
    const color = statusToColor(statusText);
    const formatter = new Intl.NumberFormat("en-US");

    return (
      <Card>
        <CardHighlight color={color} />
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
                    <StatusIcon status={statusText} inline /> {statusText} ({
                      check.status
                    })
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Last OK</DictionaryKey>
                  <DictionaryValue>
                    <RelativeDate dateTime={check.lastOK} capitalize />
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
                    <DateTime dateTime={check.issued} short />
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Ran at</DictionaryKey>
                  <DictionaryValue>
                    <KitchenTime dateTime={check.executed} />
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
