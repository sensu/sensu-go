import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import Grid from "@material-ui/core/Grid";
import ExpansionPanel from "@material-ui/core/ExpansionPanel";
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import Typography from "@material-ui/core/Typography";

import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import ErrorIcon from "@material-ui/icons/Error";

const StyledExpansionPanelSummary = withStyles(() => ({
  content: { maxWidth: "100%" },
}))(ExpansionPanelSummary);

const StyledExpansionPanelDetails = withStyles(() => ({
  root: { flexDirection: "column" },
}))(ExpansionPanelDetails);

const StyledErrorIcon = withStyles(theme => ({
  root: {
    display: "inline-block",
    verticalAlign: "-18%",
    width: "0.75em",
    height: "0.75em",
    color: theme.palette.error.main,
  },
}))(ErrorIcon);

class SilenceEntryFormPanel extends React.PureComponent {
  static propTypes = {
    title: PropTypes.node,
    summary: PropTypes.node,
    hasDefaultValue: PropTypes.bool,
    error: PropTypes.string,
    children: PropTypes.node,
  };

  static defaultProps = {
    title: undefined,
    summary: undefined,
    hasDefaultValue: false,
    error: undefined,
    children: undefined,
  };

  render() {
    const { title, summary, children, hasDefaultValue, error } = this.props;

    let summaryColor = "secondary";
    if (error) {
      summaryColor = "error";
    } else if (hasDefaultValue) {
      summaryColor = "textSecondary";
    }

    return (
      <ExpansionPanel defaultExpanded={hasDefaultValue || !!error}>
        <StyledExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
          <Grid container>
            <Grid item xs={4}>
              <Typography variant="body2" noWrap>
                {error && <StyledErrorIcon />} {title}
              </Typography>
            </Grid>
            <Grid item xs={8}>
              <Typography color={summaryColor} noWrap>
                {error || summary}
              </Typography>
            </Grid>
          </Grid>
        </StyledExpansionPanelSummary>
        <StyledExpansionPanelDetails>{children}</StyledExpansionPanelDetails>
      </ExpansionPanel>
    );
  }
}

export default SilenceEntryFormPanel;
