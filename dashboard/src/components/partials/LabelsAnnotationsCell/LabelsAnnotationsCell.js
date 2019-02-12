import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import { withStyles } from "@material-ui/core/styles";
import CodeBlock from "/components/CodeBlock";
import CodeHighlight from "/components/CodeHighlight/CodeHighlight";
import CardContent from "@material-ui/core/CardContent";
import Dictionary, {
  DictionaryKey,
  DictionaryValue,
  DictionaryEntry,
} from "/components/Dictionary";
import Grid from "@material-ui/core/Grid";
import Maybe from "/components/Maybe";
import Label from "/components/partials/Label";

const styles = () => ({
  override: {
    width: "25%",
  },
  fullWidth: {
    width: "100%",
  },
});

class LabelsAnnotationsCell extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    entity: PropTypes.object,
    check: PropTypes.object,
  };

  static defaultProps = {
    entity: null,
    check: null,
  };

  static fragments = {
    check: gql`
      fragment LabelsAnnotationsCell_check on CheckConfig {
        metadata {
          labels {
            key
            val
          }
          annotations {
            key
            val
          }
        }
      }
    `,
    entity: gql`
      fragment LabelsAnnotationsCell_entity on Entity {
        metadata {
          labels {
            key
            val
          }
          annotations {
            key
            val
          }
        }
      }
    `,
  };

  render() {
    const { check, classes, entity } = this.props;

    const object = check || entity;

    return (
      <CardContent>
        <Grid container spacing={0}>
          <Grid item xs={12} sm={12}>
            <Dictionary>
              <DictionaryEntry>
                <DictionaryKey className={classes.override}>
                  Labels
                </DictionaryKey>
                <DictionaryValue explicitRightMargin>
                  <Maybe value={object.metadata.labels} fallback="None">
                    {val =>
                      val.map(pair => [
                        <Label name={pair.key} value={pair.val} />,
                        " ",
                      ])
                    }
                  </Maybe>
                </DictionaryValue>
              </DictionaryEntry>
            </Dictionary>
          </Grid>
          <Grid item xs={12} sm={12}>
            <Dictionary>
              <DictionaryEntry>
                <DictionaryKey className={classes.override}>
                  Annotations
                </DictionaryKey>
                <DictionaryValue
                  scrollableContent
                  className={classes.fullWidth}
                >
                  {object.metadata.annotations.length > 0 ? (
                    <CodeBlock>
                      <CodeHighlight
                        language="json"
                        code={JSON.stringify(
                          object.metadata.annotations,
                          null,
                          "\t",
                        )}
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
    );
  }
}

export default withStyles(styles)(LabelsAnnotationsCell);
