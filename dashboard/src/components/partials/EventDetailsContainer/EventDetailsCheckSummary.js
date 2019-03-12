import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withStyles } from "@material-ui/core/styles";
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
import Maybe from "/components/Maybe";
import SilencedIcon from "/icons/Silence";
import Tooltip from "@material-ui/core/Tooltip";
import CronDescriptor from "/components/partials/CronDescriptor";
import NamespaceLink from "/components/util/NamespaceLink";
import InlineLink from "/components/InlineLink";
import List from "@material-ui/core/List";
import CodeBlock from "/components/CodeBlock";
import Code from "/components/Code";
import CodeHighlight from "/components/CodeHighlight/CodeHighlight";
import ListItem, { ListItemTitle } from "/components/DetailedListItem";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import ExpansionPanel from "@material-ui/core/ExpansionPanel";
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";

const styles = theme => ({
  alignmentFix: {
    boxSizing: "border-box",
  },
  fullWidth: {
    width: "100%",
  },
  expand: { color: theme.palette.text.secondary },
});

class EventDetailsCheckSummary extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object.isRequired,
    entity: PropTypes.object.isRequired,
    event: PropTypes.object.isRequired,
    classes: PropTypes.object.isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventDetailsCheckSummary_event on Event {
        isSilenced
      }
    `,
    check: gql`
      fragment EventDetailsCheckSummary_check on Check {
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
        command
        subscriptions
        stdin
        highFlapThreshold
        lowFlapThreshold
        interval
        cron
        timeout
        ttl
        totalStateChange
        roundRobin
        handlers {
          name
        }
        checkHooks {
          hooks
        }
        assets: runtimeAssets {
          id
          name
        }
        outputMetricFormat
        outputMetricHandlers {
          name
        }
      }
    `,
    entity: gql`
      fragment EventDetailsCheckSummary_entity on Entity {
        name
        namespace
      }
    `,
  };

  renderHooks() {
    const { checkHooks } = this.props.check;
    const hooks = Object.values(
      checkHooks.reduce(
        (h, list) =>
          list.hooks.reduce((j, val) => Object.assign(j, { [val]: val }), h),
        {},
      ),
    );

    return this.renderList(hooks);
  }

  renderAssets = () => {
    const { assets } = this.props.check;
    return this.renderList(assets.map(asset => asset.name));
  };

  renderList = items => {
    if (items.length === 0) {
      return "—";
    }
    return (
      <List disablePadding>
        {items.map(item => (
          <ListItem key={item}>
            <ListItemTitle>{item}</ListItemTitle>
          </ListItem>
        ))}
      </List>
    );
  };

  render() {
    const { event, check, entity, classes } = this.props;
    const statusCode = check.status;
    const status = statusCodeToId(check.status);
    const formatter = new Intl.NumberFormat("en-US");

    return (
      <Card>
        <CardHighlight color={status} />
        <CardContent>
          <Typography variant="headline" paragraph>
            Check Result
            {event.isSilenced && (
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
                <DictionaryEntry>
                  <DictionaryKey>Total State Change</DictionaryKey>
                  <DictionaryValue>
                    {entity.totalStateChange || 0}
                    {"%"}
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
                    {check.name !== "keepalive" ? (
                      <InlineLink
                        to={`/${entity.namespace}/checks/${check.name}`}
                      >
                        {check.name}
                      </InlineLink>
                    ) : (
                      check.name
                    )}
                  </DictionaryValue>
                </DictionaryEntry>
                <DictionaryEntry>
                  <DictionaryKey>Entity</DictionaryKey>
                  <DictionaryValue>
                    <NamespaceLink
                      component={InlineLink}
                      namespace={entity.namespace}
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
        <ExpansionPanel>
          <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
            <Typography variant="button" className={classes.expand}>
              Check Configuration Summary
            </Typography>
          </ExpansionPanelSummary>
          <ExpansionPanelDetails>
            <CardContent className={classes.fullWidth}>
              <Grid container spacing={0}>
                <Grid item xs={12} sm={6}>
                  <Dictionary>
                    <DictionaryEntry>
                      <DictionaryKey>Check</DictionaryKey>
                      <DictionaryValue>
                        {check.name !== "keepalive" ? (
                          <InlineLink
                            to={`/${entity.namespace}/checks/${check.name}`}
                          >
                            {check.name}
                          </InlineLink>
                        ) : (
                          check.name
                        )}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>STDIN</DictionaryKey>
                      <DictionaryValue>
                        <CodeHighlight
                          language="bash"
                          component={Code}
                          code={check.stdin || "false"}
                        />
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Subscriptions</DictionaryKey>
                      <DictionaryValue>
                        {check.subscriptions.length > 0 ? (
                          <List disablePadding>
                            {check.subscriptions.map(subscription => (
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
                    <DictionaryEntry>
                      <DictionaryKey>Timeout</DictionaryKey>
                      <DictionaryValue>
                        <Maybe value={check.timeout} fallback="Never">
                          {timeout => `${timeout}s`}
                        </Maybe>
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>TTL</DictionaryKey>
                      <DictionaryValue>
                        <Maybe value={check.ttl} fallback="Forever">
                          {ttl => `${ttl}s`}
                        </Maybe>
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Handlers</DictionaryKey>
                      <DictionaryValue>
                        {check.handlers.length > 0 ? (
                          <List disablePadding>
                            {check.handlers.map(handler => (
                              <ListItem key={handler.name}>
                                <ListItemTitle>{handler.name}</ListItemTitle>
                              </ListItem>
                            ))}
                          </List>
                        ) : (
                          "—"
                        )}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Hooks</DictionaryKey>
                      <DictionaryValue>{this.renderHooks()}</DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Assets</DictionaryKey>
                      <DictionaryValue>{this.renderAssets()}</DictionaryValue>
                    </DictionaryEntry>
                  </Dictionary>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Dictionary>
                    <DictionaryEntry>
                      <DictionaryKey>Command</DictionaryKey>
                      <DictionaryValue scrollableContent>
                        {check.command ? (
                          <CodeBlock>
                            <CodeHighlight
                              language="bash"
                              code={check.command}
                              component="code"
                            />
                          </CodeBlock>
                        ) : (
                          "—"
                        )}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Schedule</DictionaryKey>
                      <DictionaryValue>
                        <Maybe
                          value={check.cron}
                          fallback={`${check.interval}s`}
                        >
                          {cron => (
                            <CronDescriptor capitalize expression={cron} />
                          )}
                        </Maybe>
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Round Robin</DictionaryKey>
                      <DictionaryValue>
                        {check.roundRobin ? "Yes" : "No"}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Flap Threshold</DictionaryKey>
                      <DictionaryValue>
                        High: {check.highFlapThreshold} Low:{" "}
                        {check.lowFlapThreshold}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Accepts STDIN?</DictionaryKey>
                      <DictionaryValue>
                        {check.stdin ? "Yes" : "No"}
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Metric Format</DictionaryKey>
                      <DictionaryValue>
                        <Maybe
                          value={check.outputMetricFormat}
                          fallback="None"
                        />
                      </DictionaryValue>
                    </DictionaryEntry>
                    <DictionaryEntry>
                      <DictionaryKey>Metric Handlers</DictionaryKey>
                      <DictionaryValue>
                        {check.outputMetricHandlers.length > 0 ? (
                          <List disablePadding>
                            {check.outputMetricHandlers.map(handler => (
                              <ListItem key={handler.name}>
                                <ListItemTitle>{handler.name}</ListItemTitle>
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
          </ExpansionPanelDetails>
        </ExpansionPanel>
      </Card>
    );
  }
}

export default withStyles(styles)(EventDetailsCheckSummary);
