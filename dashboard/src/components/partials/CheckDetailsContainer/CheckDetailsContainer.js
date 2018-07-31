import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Code from "/components/Code";
import CollapsingMenu from "/components/partials/CollapsingMenu";
import Content from "/components/Content";
import CronDescriptor from "/components/partials/CronDescriptor";
import DeleteIcon from "@material-ui/icons/Delete";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import Divider from "@material-ui/core/Divider";
import Grid from "@material-ui/core/Grid";
import List from "@material-ui/core/List";
import ListItem, { ListItemTitle } from "/components/DetailedListItem";
import LiveIcon from "/icons/Live";
import Loader from "/components/util/Loader";
import Maybe from "/components/Maybe";
import Monospaced from "/components/Monospaced";
import SilencedIcon from "/icons/Silence";
import Typography from "@material-ui/core/Typography";
import Tooltip from "@material-ui/core/Tooltip";
import QueueIcon from "@material-ui/icons/Queue";

import DeleteAction from "./CheckDetailsDeleteAction";
import ExecuteAction from "./CheckDetailsExecuteAction";
import UnsilenceAction from "./CheckDetailsUnsilenceAction";

class CheckDetailsContainer extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
    loading: PropTypes.bool.isRequired,
    poller: PropTypes.object.isRequired,
    refetch: PropTypes.func,
  };

  static defaultProps = {
    check: null,
    refetch: () => null,
  };

  static fragments = {
    checkConfig: gql`
      fragment CheckDetailsContainer_checkConfig on CheckConfig {
        deleted @client
        id
        name
        command
        subscriptions
        stdin
        highFlapThreshold
        lowFlapThreshold
        isSilenced

        interval
        cron
        timeout
        ttl
        publish
        roundRobin
        handlers {
          name
        }

        outputMetricFormat
        outputMetricHandlers {
          name
        }

        checkHooks {
          hooks
        }

        # proxyEntityId
        proxyRequests {
          entityAttributes
          splay
        }

        envVars
        extendedAttributes

        ...CheckDetailsDeleteAction_check
        ...CheckDetailsExecuteAction_check
        ...CheckDetailsUnsilenceAction_check
      }

      ${DeleteAction.fragments.check}
      ${ExecuteAction.fragments.check}
      ${UnsilenceAction.fragments.check}
    `,
  };

  renderSchedule() {
    const { interval, cron } = this.props.check;

    if (interval > 0) {
      return `Every ${interval}s`;
    } else if (cron && cron.length > 0) {
      return <CronDescriptor expression={cron} />;
    }
    return "Never";
  }

  renderHooks() {
    const { checkHooks } = this.props.check;
    const hooks = Object.values(
      checkHooks.reduce(
        (h, list) =>
          list.hooks.reduce((j, val) => Object.assign(j, { [val]: val }), h),
        {},
      ),
    );

    if (hooks.length === 0) {
      return "—";
    }

    return (
      <List disablePadding>
        {hooks.map(hook => (
          <ListItem key={hook}>
            <ListItemTitle>{hook}</ListItemTitle>
          </ListItem>
        ))}
      </List>
    );
  }

  render() {
    const { check, loading, poller, refetch } = this.props;

    return (
      <Loader loading={loading} passthrough>
        {check && (
          <React.Fragment>
            <Content bottomMargin>
              <div style={{ flexGrow: 1 }} />
              <CollapsingMenu>
                {check.isSilenced && (
                  <UnsilenceAction check={check} refetch={refetch}>
                    {unsilence => (
                      <CollapsingMenu.Button
                        title="Unsilence"
                        icon={<SilencedIcon />}
                        onClick={unsilence}
                      />
                    )}
                  </UnsilenceAction>
                )}
                <ExecuteAction check={check}>
                  {executeCheck => (
                    <CollapsingMenu.Button
                      title="Execute"
                      icon={<QueueIcon />}
                      onClick={executeCheck}
                    />
                  )}
                </ExecuteAction>
                <DeleteAction check={check}>
                  {deleteCheck => (
                    <CollapsingMenu.Button
                      title="Delete"
                      icon={<DeleteIcon />}
                      onClick={deleteCheck}
                    />
                  )}
                </DeleteAction>
                <CollapsingMenu.Button
                  pinned
                  title="LIVE"
                  icon={<LiveIcon active={poller.running} />}
                  onClick={() =>
                    poller.running ? poller.stop() : poller.start()
                  }
                />
              </CollapsingMenu>
            </Content>
            <Content>
              <Grid container spacing={16}>
                <Grid item xs={12}>
                  <Card>
                    <CardContent>
                      <Typography variant="headline">
                        {check.isSilenced > 0 && (
                          <Tooltip title="Silenced">
                            <SilencedIcon style={{ float: "right" }} />
                          </Tooltip>
                        )}
                        Configuration
                      </Typography>
                      <Typography variant="caption" paragraph>
                        Defines when, where and how a check is executed.
                      </Typography>
                      <Grid container spacing={0}>
                        <Grid item xs={12} sm={6}>
                          <Dictionary>
                            <DictionaryEntry>
                              <DictionaryKey>Name</DictionaryKey>
                              <DictionaryValue>{check.name}</DictionaryValue>
                            </DictionaryEntry>

                            <DictionaryEntry>
                              <DictionaryKey>Command</DictionaryKey>
                              <DictionaryValue>
                                <Code>{check.command}</Code>
                              </DictionaryValue>
                            </DictionaryEntry>

                            <DictionaryEntry>
                              <DictionaryKey>Subscriptions</DictionaryKey>
                              <DictionaryValue>
                                {check.subscriptions.length > 0 ? (
                                  <List disablePadding>
                                    {check.subscriptions.map(subscription => (
                                      <ListItem key={subscription}>
                                        <ListItemTitle>
                                          {subscription}
                                        </ListItemTitle>
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
                              <DictionaryKey>Published?</DictionaryKey>
                              <DictionaryValue>
                                {check.publish ? "Yes" : "No"}
                              </DictionaryValue>
                            </DictionaryEntry>

                            <DictionaryEntry>
                              <DictionaryKey>Schedule</DictionaryKey>
                              <DictionaryValue>
                                {this.renderSchedule()}
                              </DictionaryValue>
                            </DictionaryEntry>

                            <DictionaryEntry>
                              <DictionaryKey>Round Robin</DictionaryKey>
                              <DictionaryValue>
                                {check.roundRobin ? "Yes" : "No"}
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
                              <DictionaryKey>Proxy Entity ID</DictionaryKey>
                              <DictionaryValue>
                                <Maybe
                                  value={check.outputMetricFormat}
                                  fallback="None"
                                />
                              </DictionaryValue>
                            </DictionaryEntry>

                            <DictionaryEntry>
                              <DictionaryKey>Proxy Requests</DictionaryKey>
                              <DictionaryValue>
                                <Maybe
                                  value={check.proxyRequests}
                                  fallback="None"
                                >
                                  {val => JSON.stringify(val)}
                                </Maybe>
                              </DictionaryValue>
                            </DictionaryEntry>
                          </Dictionary>
                        </Grid>

                        <Grid item xs={12} sm={6}>
                          <Dictionary>
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
                              <DictionaryKey>ENV Vars</DictionaryKey>
                              <DictionaryValue>
                                {check.envVars.length > 0 ? (
                                  <Monospaced highlight background>
                                    {check.envVars.join("\n")}
                                  </Monospaced>
                                ) : (
                                  "None"
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
                              <DictionaryKey>Handlers</DictionaryKey>
                              <DictionaryValue>
                                {check.handlers.length > 0 ? (
                                  <List disablePadding>
                                    {check.handlers.map(handler => (
                                      <ListItem key={handler.name}>
                                        <ListItemTitle>
                                          {handler.name}
                                        </ListItemTitle>
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
                              <DictionaryValue>
                                {this.renderHooks()}
                              </DictionaryValue>
                            </DictionaryEntry>
                          </Dictionary>
                        </Grid>
                        <Grid item xs={12} sm={6}>
                          <Dictionary>
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
                                        <ListItemTitle>
                                          {handler.name}
                                        </ListItemTitle>
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

                    {check.extendedAttributes && (
                      <React.Fragment>
                        <Divider />
                        <Monospaced background>
                          <CardContent>
                            {`# Extra\n\n${JSON.stringify(
                              check.extendedAttributes,
                              null,
                              "\t",
                            )}`}
                          </CardContent>
                        </Monospaced>
                      </React.Fragment>
                    )}
                  </Card>
                </Grid>
              </Grid>
            </Content>
          </React.Fragment>
        )}
      </Loader>
    );
  }
}

export default CheckDetailsContainer;
