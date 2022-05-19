import * as React from 'react';
import { useRef, useState, useEffect } from 'react';
import Countdown from 'react-countdown';
import { StyledEngineProvider, ThemeProvider, createTheme } from '@mui/material/styles';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import TextareaAutosize from '@mui/material/TextareaAutosize';
import './style.css';
import logo from './logo.txt';
import git_image from './git.png';
import commit_image from './commit.png';
import loading_image from './loading.gif';
import { isTemplateSpan } from 'typescript';

type User = {
    user: string,
    lastSHA1: string,
    reviewcommits?: string[]
}

type Repo = {
    provider: string,
    url: string,
    branch: string,
    lastSHA1: string,
    users?: User[],
    reviewcommits?: string[]
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;

    return (
        <div
            role="tabpanel"
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}
        >
            {value === index && (
            <Box sx={{ p: 3 }}>
                {children}
            </Box>
            )}
        </div>
    );
}

function a11yProps(index: number) {
    return {
        id: `simple-tab-${index}`,
        'aria-controls': `simple-tabpanel-${index}`,
    };
}

function Commits() {
    const [result, setRepoResult] = useState<Repo[]>([]);

    useEffect(() => {
        const get_repo_diff = async () => {
            const data = await fetch("http://localho.st:1323/config", {
                method: "GET"
            });
            const jsonData = await data.json();
            setRepoResult(jsonData.repos);
        };
    
        get_repo_diff();
    }, []);

    const reviewedCommit = async(event) => {
        event.preventDefault();
        console.log(event.target.id);

        const review_commit = async () => {
            const data = await fetch("http://localho.st:1323/review/" + event.target.id, {
                method: "GET"
            });
            const jsonData = await data.json();
            setRepoResult(jsonData.repos);
            console.log(jsonData);
        };

        review_commit();
    }

    const RepoUserCommits = (props) => {
        return (
            <li key={Math.random()}>
            {
                props.user.reviewcommits?.map((user_commit, user_commit_index, user_reviewcommits) => {
                    let [user_title, user_url] = user_commit.split(",#,");
                    let [user_commit_date, user_commit_title, commit_user] = user_title.split("||");
                    let user_commit_ref = props.index.toString() + "_" + props.user_index.toString() + "_" +  user_commit_index.toString();

                    return (
                        <div key={Math.random()}>
                            <span>
                                <img src={commit_image} alt="commit image" width="14" height="14" className="commit_image"/>
                                <a target="_blank" title={user_commit_title} href={user_url}  className="linkCommit">{user_commit_title}</a>
                                <br/>
                                <span className="commitByLine">on <b className="commitByLineDate">{user_commit_date}</b> by <b className="commitByLineUser">{commit_user}</b></span>
                            </span>
                            <span className="review_span">
                            {
                                user_reviewcommits.length - 1 === user_commit_index ? <Button variant="outlined" className='review_button' id={user_commit_ref} onClick={reviewedCommit}>Reviewed</Button> : <span className='review_skip'>...</span>
                            }
                            </span>
                        </div>
                    )
                })
            }
            </li>
        )
    }

    const RepoUsers = (props) => {
        return (
            <ol key={Math.random()} className="user_commit_list">
            {
                props.users?.map((user, user_index, users) => {
                    return <RepoUserCommits key={Math.random()} index={props.index} user={user} user_index={user_index} users={users} />
                })
            }
            </ol>
        )
    }

    const RepoCommits = (props) => {
        return (
            <ol key={Math.random()} className="commit_list">
            {                                                
                props.item.reviewcommits?.map((commit, commit_index, reviewcommits) => {
                    let [title, url] = commit.split(",#,");
                    let [date, commit_title, user] = title.split("||");
                    let commit_ref = props.index.toString() + "_" + commit_index.toString();
                    return (
                        <li key={Math.random()}> 
                            <span>                                           
                                <img src={commit_image} alt="commit image" width="14" height="14" className="commit_image"/>
                                <a target="_blank" title={commit_title} href={url}  className="linkCommit">{commit_title}</a>
                                <br/>
                                <span className="commitByLine">on <b className="commitByLineDate">{date}</b> by <b className="commitByLineUser">{user}</b></span>
                            </span>
                            <span className="review_span">
                            {
                                reviewcommits.length - 1 === commit_index ? 
                                    <Button variant="outlined" className='review_button' id={commit_ref} onClick={reviewedCommit}>Reviewed</Button> 
                                : <span className='review_skip'>...</span>
                            }
                            </span>
                        </li>
                    )
                })
            }
            </ol> 
        )
    }

    return (
        <>
        <ol key={Math.random()} className="repo_list">
        {
            result.map((item, index) => 
                <li key={Math.random()}><img src={git_image} alt="git logo" width="16" height="16" className="git_image"/>
                    <a target="_blank" title={item.url} href={item.url} className="linkRepo">{item.url}</a>
                    {
                        item.reviewcommits !== null ? <RepoCommits key={Math.random()} index={index} item={item} /> : <RepoUsers key={Math.random()} index={index} users={item.users} />
                    }
                </li>
            )                            
        }
        </ol>
        </>
    );
}


function Config() {
    const [pretty, setPrettyResult] = useState<string>();
    const [loading, setloadingResult] = useState<boolean>();
    const formRef = useRef()

    useEffect(() => {
        const get_config = async () => {
            const data = await fetch("http://localho.st:1323/pretty", {
                method: "GET"
            });
            const text = await data.text();
            setPrettyResult(text);
        };
    
        get_config();
    }, []);

    const saveConfig = async(event) => {   
        setloadingResult(true);
        event.preventDefault();
        const data = new FormData(formRef.current)
        fetch('http://localho.st:1323/update', {
          method: 'POST', 
          mode: 'cors', 
          body: data
        }).then(response => { 
            setloadingResult(false);
            console.log(response);
        });     
    }

    return (
        <>
        <form ref={formRef} onSubmit={saveConfig}>
            <Button variant="contained" type="submit">
            {
                loading ? <img src={loading_image} alt="commit image" width="16" height="16" id="saving_image" /> : "Save"
            }
            </Button>
            <br/><br/>
            <TextareaAutosize
                aria-label="empty textarea"
                placeholder="Empty"
                defaultValue={pretty}
                className="config"
                name="json_config"
            />
        </form>
        </>
    );
}

function ConfigState() {
    const [stateConfig, setStateResult] = useState<string>();

    useEffect(() => {
        const get_state = async () => {
            const data = await fetch("http://localho.st:1323/state", {
                method: "GET"
            });
            const text = await data.text();
            setStateResult(text);
        };
    
        get_state();
    }, []);

    return (
        <>
        <pre name="state_config" className="configState">{stateConfig}</pre>
        </>
    );
}

const renderer = ({ minutes, seconds, completed }) => {
    if (completed) {
        window.location.reload();
    } else {
        return <div className='countdown'>Updating in {minutes}:{seconds}</div>;
    }
};

const darkTheme = createTheme({
    palette: {
      mode: 'dark',
    },
});

export const Layout = () => {
    const [value, setValue] = React.useState(0);

    const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
        setValue(newValue);
    };

    return (
        <div className="main_wrapper"><pre className='logo'>{logo}</pre>
            <Countdown date={Date.now() + 900000} renderer={renderer}/>
            <ThemeProvider theme={darkTheme}>
                <StyledEngineProvider injectFirst>
                    <Box sx={{ width: '100%' }}>
                        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                        <Tabs value={value} onChange={handleChange} aria-label="basic tabs example">
                            <Tab label="Latest commits" {...a11yProps(0)} />
                            <Tab label="Configuration" {...a11yProps(1)} />
                            <Tab label="State" {...a11yProps(2)} />
                        </Tabs>
                        </Box>
                        <TabPanel value={value} index={0}>
                            <Commits />
                        </TabPanel>
                        <TabPanel value={value} index={1}>
                            <Config />
                        </TabPanel>
                        <TabPanel value={value} index={2}>
                            <ConfigState />
                        </TabPanel>
                    </Box>
                </StyledEngineProvider>
            </ThemeProvider>
        </div>
    );
}
