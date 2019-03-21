import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Divider from "@material-ui/core/Divider";
import Grid from "@material-ui/core/Grid";
import List from "@material-ui/core/List";
import Typography from "@material-ui/core/Typography";
import Tooltip from "@material-ui/core/Tooltip";

import Duration from "/lib/component/base/Duration";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/lib/component/base/Dictionary";
import ListItem, { ListItemTitle } from "/lib/component/base/DetailedListItem";
import Maybe from "/lib/component/util/Maybe";
import CodeBlock from "/lib/component/base/CodeBlock";
import CodeHighlight from "/lib/component/base/CodeHighlight";
import SilencedIcon from "/lib/component/icon/Silence";

import CronDescriptor from "/app/component/partial/CronDescriptor";
import LabelsAnnotationsCell from "/app/component/partial/LabelsAnnotationsCell";

class CheckDetailsConfiguration extends React.PureComponent {
  static propTypes = {
    check: PropTypes.object,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsConfiguration_check on CheckConfig {
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

        proxyEntityName
        proxyRequests {
          splay
          splayCoverage
          entityAttributes
        }

        assets: runtimeAssets {
          id
          name
        }

        envVars

        metadata {
          ...LabelsAnnotationsCell_objectmeta
        }
      }
      ${LabelsAnnotationsCell.fragments.objectmeta}
    `,
  };

  renderSchedule() {
    const { interval, cron } = this.props.check;

    if (interval > 0) {
      return (
        <React.Fragment>
          Every <Duration duration={interval * 1000} />
        </React.Fragment>
      );
    } else if (cron && cron.length > 0) {
      return <CronDescriptor capitalize expression={cron} />;
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
    const { check } = this.props;

    return (
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
                  <DictionaryKey>Published?</DictionaryKey>
                  <DictionaryValue>
                    {check.publish ? "Yes" : "No"}
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
                  <DictionaryValue>{this.renderSchedule()}</DictionaryValue>
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
                      {val => <Duration duration={val * 1000} />}
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
                    <Maybe value={check.proxyEntityName} fallback="None" />
                  </DictionaryValue>
                </DictionaryEntry>

                <DictionaryEntry fullWidth={!!check.proxyRequests}>
                  <DictionaryKey>Proxy Requests</DictionaryKey>
                  <DictionaryValue explicitRightMargin>
                    <Maybe value={check.proxyRequests} fallback="None">
                      {val => (
                        <CodeBlock>
                          <CodeHighlight
                            language="json"
                            code={JSON.stringify(
                              {
                                entity_attributes: val.entityAttributes,
                                splay: val.splay,
                                splay_coverage: val.splayCoverage,
                              },
                              null,
                              "  ",
                            )}
                            component="code"
                          />
                        </CodeBlock>
                      )}
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

                <DictionaryEntry fullWidth={check.envVars.length > 0}>
                  <DictionaryKey>ENV Vars</DictionaryKey>
                  <DictionaryValue>
                    {check.envVars.length > 0 ? (
                      <CodeBlock>
                        <CodeHighlight
                          language="properties"
                          code={check.envVars.join("\n")}
                          component="code"
                        />
                      </CodeBlock>
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
                  <DictionaryKey>Metric Format</DictionaryKey>
                  <DictionaryValue>
                    <Maybe value={check.outputMetricFormat} fallback="None" />
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

        <Divider />
        <LabelsAnnotationsCell check={check} />
      </Card>
    );
  }
}

export default CheckDetailsConfiguration;
