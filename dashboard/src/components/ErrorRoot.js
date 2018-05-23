import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import StackTrace from "stacktrace-js";
import ErrorStackParser from "error-stack-parser";
import { isApolloError } from "apollo-client/errors/ApolloError";

import Dialog from "@material-ui/core/Dialog";
import DialogContent from "@material-ui/core/DialogContent";
import DialogTitle from "@material-ui/core/DialogTitle";

import ErrorIcon from "@material-ui/icons/Error";

import { sourceURL, sourceRevision } from "/buildInfo.macro";

import FetchError from "/errors/FetchError";
import ReactError from "/errors/ReactError";

const Title = withStyles(theme => ({
  root: {
    background: theme.palette.error.main,
    display: "flex",
    alignItems: "center",
    color: theme.palette.error.contrastText,
  },
}))(props => <DialogTitle {...props} disableTypography />);

const Icon = withStyles(theme => ({
  root: {
    width: 32,
    height: 32,
    marginRight: theme.spacing.unit * 2,
  },
}))(ErrorIcon);

const Pre = withStyles(() => ({
  root: {
    marginTop: 16,
    lineHeight: 1.5,
    fontSize: 14,
    fontFamily: `"SFMono-Regular", Consolas, "Liberation Mono", Menlo,Courier, monospace`,
  },
}))(({ classes, ...props }) => <pre {...props} className={classes.root} />);

const formatFrame = ({ functionName, fileName, lineNumber }) => ({
  functionName,
  url: /^src\//.test(fileName)
    ? `${sourceURL}blob/${sourceRevision}/dashboard/${fileName}#L${lineNumber}`
    : undefined,
  source: `${fileName}:${lineNumber}`,
});

const issueLink = ({ name, message, frames, componentStack }) => {
  const title = `[Web UI] ${name}: ${message}`;

  const body = `
<!--- Provide a general summary of the issue in the Title above -->

## Expected Behavior
<!--- If you're looking for help, please see https://sensuapp.org/support for resources --->
<!--- If you're describing a bug, tell us what should happen -->
<!--- If you're suggesting a change/improvement, tell us how it should work -->

## Current Behavior
<!--- If describing a bug, tell us what happens instead of the expected behavior -->
<!--- If suggesting a change/improvement, explain the difference from current behavior -->

## Possible Solution
<!--- Not obligatory, but suggest a fix/reason for the bug, -->
<!--- or ideas as to the implementation of the addition or change -->

## Steps to Reproduce (for bugs)
<!--- Provide a link to a live example, or an unambiguous set of steps to -->
<!--- reproduce this bug. Include code or configuration to reproduce, if relevant -->
1.
2.
3.
4.

## Context
<!--- How has this issue affected you? What are you trying to accomplish? -->
<!--- Providing context (e.g. links to configuration settings, stack strace or log data) helps us come up with a solution that is most useful in the real world -->

**Web UI Stack Trace:**

${frames
    .slice(0, 15)
    .map(
      ({ functionName, url, source }) =>
        `at \`${functionName}\` (${url ? `[${source}](${url})` : source})`,
    )
    .join("\n")}
${
    componentStack instanceof ReactError
      ? `


Component Stack:

${"```"}${componentStack.replace(/^ {4}/gm, "")}
${"```"}`
      : ""
  }

## Your Environment
<!--- Include as many relevant details about the environment you experienced the bug in -->
* Sensu version used (sensuctl, sensu-backend, and/or sensu-agent): \`${sourceRevision}\`
* Installation method (packages, binaries, docker etc.):
* Operating System and version (e.g. Ubuntu 14.04):`.slice(0);

  const params = `?title=${encodeURIComponent(title)}&body=${encodeURIComponent(
    body,
  )}`;

  return `${sourceURL}issues/new${params}`;
};

const unwrapError = error => {
  const meta = {};

  if (error instanceof ReactError) {
    return {
      componentStack: error.componentStack,
      ...unwrapError(error.original),
    };
  }

  if (isApolloError(error) && error.networkError) {
    return unwrapError(error.networkError);
  }

  if (error instanceof FetchError) {
    if (error.original) {
      return {
        url: error.url,
        statusCode: error.statusCode,
        ...unwrapError(error.original),
      };
    }

    meta.url = error.url;
    meta.statusCode = error.statusCode;
  }

  return {
    ...meta,
    name: error.name,
    stack: error.stack,
    message: error.message,
  };
};

class ErrorRoot extends React.PureComponent {
  static propTypes = {
    // eslint-disable-next-line react/no-unused-prop-types
    error: PropTypes.instanceOf(Error).isRequired,
  };

  state = {};

  static getDerivedStateFromProps(props) {
    const unwrapped = unwrapError(props.error);
    return {
      ...unwrapped,
      frames: ErrorStackParser.parse(unwrapped)
        .map(frame => ({
          ...frame,
          functionName: `${frame.functionName}`,
          fileName: frame.fileName.replace(window.location.origin, ""),
        }))
        .map(formatFrame),
    };
  }

  componentDidMount() {
    StackTrace.fromError(this.state).then(frames => {
      this.setState({
        frames: frames
          .map((frame, i) => ({
            ...frame,
            functionName: this.state.frames[i].functionName,
            fileName: frame.fileName.replace(window.location.origin, ""),
          }))
          .map(formatFrame),
      });
    });
  }

  render() {
    const {
      stack,
      frames,
      name,
      message,
      url,
      statusCode,
      componentStack,
    } = this.state;

    return (
      <Dialog open>
        <Title>
          <Icon />
          <Typography variant="title" color="inherit">
            Something went wrong
          </Typography>
        </Title>
        <DialogContent>
          <Pre>
            <strong>
              {name}: {message}
            </strong>
            {"\n\n"}
            Environment:
            {"\n"}
            source revision: {sourceRevision}
            {"\n\n"}
            Stack Trace:
            {frames ? (
              <div>
                {frames.map((frame, index) => (
                  // eslint-disable-next-line react/no-array-index-key
                  <div key={index}>
                    {"   "}at {frame.functionName} ({frame.url ? (
                      <a href={frame.url} target="_blank">
                        {frame.source}
                      </a>
                    ) : (
                      frame.source
                    )})
                  </div>
                ))}
              </div>
            ) : (
              <div>{stack}</div>
            )}
            {componentStack && (
              <div>
                {"\n"}
                Component Stack:
                {"\n"}
                {componentStack.replace(/^ {4}/gm, "  ")}
              </div>
            )}
            {url && (
              <div>
                {"\n"}
                Fetch URL:
                {"\n"}
                {url.replace(window.location.origin, "")}
              </div>
            )}
            {statusCode !== undefined && (
              <div>
                {"\n"}
                Fetch Status:
                {"\n"}
                {statusCode}
              </div>
            )}
            {"\n\n"}
            <a href={issueLink(this.state)} target="_blank">
              submit new issue
            </a>{" "}
            (populated with current error details)
          </Pre>
        </DialogContent>
      </Dialog>
    );
  }
}

export default ErrorRoot;
