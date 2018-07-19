import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Card from "@material-ui/core/Card";
import CardContent from "@material-ui/core/CardContent";
import Grid from "@material-ui/core/Grid";
import List from "@material-ui/core/List";
import Typography from "@material-ui/core/Typography";

import Loader from "/components/util/Loader";

import ButtonSet from "/components/ButtonSet";
import Code from "/components/Code";
import Content from "/components/Content";
import ListItem, {
  ListItemTitle,
  ListItemSubtitle,
} from "/components/DetailedListItem";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";

import DeleteAction from "./CheckDetailsDeleteAction";

class CheckDetailsContainer extends React.PureComponent {
  static propTypes = {
    client: PropTypes.object.isRequired,
    check: PropTypes.object,
    loading: PropTypes.bool.isRequired,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    checkConfig: gql`
      fragment CheckDetailsContainer_checkConfig on CheckConfig {
        deleted @client
        id
        name
        command
        handlers {
          id
          type
          command
        }
        interval
        subscriptions
        timeout
        ttl

        ...CheckDetailsDeleteAction_checkConfig
      }

      ${DeleteAction.fragments.checkConfig}
    `,
  };

  state = {
    pendingRequests: 0,
  };

  handleRequestStart = () => {
    this.setState(({ pendingRequests }) => ({
      pendingRequests: pendingRequests + 1,
    }));
  };

  handleRequestEnd = () => {
    this.setState(({ pendingRequests }) => ({
      pendingRequests: pendingRequests - 1,
    }));
  };

  render() {
    const { client, check, loading } = this.props;
    const { pendingRequests } = this.state;
    const hasPendingRequests = pendingRequests > 0;

    return (
      <Loader loading={loading || hasPendingRequests} passthrough>
        {check && (
          <React.Fragment>
            <Content bottomMargin>
              <div style={{ flexGrow: 1 }} />
              <ButtonSet>
                <DeleteAction
                  client={client}
                  check={check}
                  onRequestStart={this.handleRequestStart}
                  onRequestEnd={this.handleRequestEnd}
                />
              </ButtonSet>
            </Content>
            <Content>
              <Grid container spacing={16}>
                <Grid item xs={12}>
                  <Card>
                    <CardContent>
                      <Typography variant="headline" paragraph>
                        Check Configuration
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
                              <DictionaryKey>Interval</DictionaryKey>
                              <DictionaryValue>
                                {check.interval}
                              </DictionaryValue>
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
                        </Grid>
                      </Grid>
                    </CardContent>
                  </Card>
                </Grid>

                <Grid item xs={12} sm={6}>
                  <Card>
                    <CardContent>
                      <Typography variant="headline" paragraph>
                        Subscriptions
                      </Typography>
                      {check.subscriptions.length ? (
                        <List disablePadding>
                          {check.subscriptions.map(subscription => (
                            <ListItem key={subscription}>
                              <ListItemTitle>{subscription}</ListItemTitle>
                            </ListItem>
                          ))}
                        </List>
                      ) : (
                        <Typography variant="caption" paragraph>
                          no subscriptions
                        </Typography>
                      )}
                    </CardContent>
                  </Card>
                </Grid>

                <Grid item xs={12} sm={6}>
                  <Card>
                    <CardContent>
                      <Typography variant="headline" paragraph>
                        Handlers
                      </Typography>
                      {check.handlers.length ? (
                        <List disablePadding>
                          {check.handlers.map(handler => (
                            <ListItem key={handler.id}>
                              <ListItemTitle>{handler.name}</ListItemTitle>
                              <ListItemSubtitle>
                                <Code>{handler.command}</Code>
                              </ListItemSubtitle>
                            </ListItem>
                          ))}
                        </List>
                      ) : (
                        <Typography variant="caption" paragraph>
                          no handlers
                        </Typography>
                      )}
                    </CardContent>
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
